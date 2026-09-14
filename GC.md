# Garbage Collection

When our project started it used a simple bump-allocator.  That just meant reserving a huge contiguous chunk of memory and maintaining a simple count of memory used.  Every request would just advance the "used" pointer by the size requested, and return the previous value.  This was fast and simple, but meant there was no possibility to free memory.

The introduction of the [inception](INCEPTION.md) lisp-interpreter, and to a lesser extent our brainfuck and nqueens programs, really made it apparent that this wasn't tenable; large programs would exhaust the available heap with items that were no longer referenced.

To start with I did the obvious thing and simply made the allocation region larger, ignoring the problem.  But eventually that too became untenable.  So now I've implemented a stop & copy garbage collector, using Cheney's algorithm.   Every time our `(cons ..)` primitive is called we check to see if we should invoke the garbage-collection process.

The `(cons ..)` primitive is a lisp-fundamental, so I figure that is going to be called pretty often in user-programs, either directly or via the `(list ...)` wrapper.  But if that isn't the case you may also trigger the garbage-collection process explicitly, and see the stats via these methods:

* `(sys-gc)` - run the garbage collection process immediately.
* `(sys-gc-threshold)` - We run the GC process after a fixed number of allocations have been made, this retrieves the value of that threshold.
  * Change it via the `GC_ALLOC_THRESHOLD` environmental variable.
* `(sys-heap-allocs)` -  Return the number of memory allocations made.
* `(sys-heap-bytes)` - Return the size of the heap.
* `(sys-heap-data)` - Return the contents of the heap as a list of entries.
* `(sys-heap-dump)` - Dump a summary of the heap to the console.
  * This is implemented in assembly and literally writes to STDOUT.  There is no control over the formatting.
  * Use `sys-heap-data` if you want to get the heap-data and format it for printing yourself.
* `(sys-heap-objects)` - Return the number of objects stored upon the heap.



## Implementation Overview

The stop and copy implementation is pretty simple:

* We have *two* heap areas, each of which are an identical size.
* One heap is used as the backing-store for all allocations we make.
* When a `sys-gc` request is made the current heap is inspected and all live items are copied to the other heap.
  * The new heap is then made the active one, which essentially orphans and frees the unreachable entries upon the old heap.
  * As a nice side-effect this removes any fragmentation, there are no holes in the new heap.  Allocation continues to grow the heap with no need to worry about using previously-freed objects.
* The copying process has to deal with global variables, objects held within stack-frames, and those objects which might be held inside registers.
  * For register contents we cheat a little.
     * The `(cons ..)` primitive is the only one that is used to trigger "auto GC",  and we know `cons` can only be called with two arguments, so we only have to consider the two registers RDI & RSI.

TLDR; Our roots are "globals", "stack-locals", and potentially the contents of the two AMD64 processor registers `rdi` and `rsi`.



## Changing Limits

By default the garbage collection process is triggered when more than 100,000 allocations have been made, but if you wish you can adjust the threshold by setting `GC_ALLOC_THRESHOLD` environmental variable before you execute one of the compiled binaries:

    GC_ALLOC_THRESHOLD=1      ./foo
    GC_ALLOC_THRESHOLD=500000 ./foo

Setting the threshold low will ensure the heap is as small as possible, at a cost that the garbage collector will run more frequently.  Setting it to a "mid-high" value is perhaps more efficient, the garbage collector will run less often, and do more work each time.  But chances are that time won't dominate the _total_ runtime, which is a risk if it is set too low.



## GC Reports

The compiler will invoke the function `at_exit` after calling the `main` function, and we ship a stub version of that function which reports a summary of memory usage if the `GC_DUMP_STATS` environmental variable is non-empty.

You can override that function to do other things, if you wish, or just use it for diagnostics:

     $ GC_DUMP_STATS=1 ./inception --help
     Loaded stdlib.lisp in 1770ms.
     heap objects:48192
     heap size:2M allocation-limit:100000
     allocations:237491

If you force the limit to be low you'll see there are fewer live objects on the heap at the time the process terminates (but you'll also see that it took over 6 seconds to load/process the standard library, due to the overhead of the constant GC process):

     $ GC_ALLOC_THRESHOLD=1  GC_DUMP_STATS=1 ./inception --help
     Loaded stdlib.lisp in 6311ms.
     heap objects:11106
     heap size:543K allocation-limit:1
     allocations:237496
