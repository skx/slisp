# Garbage Collection

When our project started it used a simple bump-allocator.  That just meant reserving a huge contiguous chunk of memory and maintaining a simple count of memory used.  Every request would just advance the "used" pointer by the size requested, and return the previous value.  This was fast and simple, but meant there was no possibility to free memory.

The introduction of the `inception` lisp-interpreter, and to a lesser extent our brainfuck and nqueens programs, really made it apparent that this wasn't tenable.  Large programs would exhaust the available heap with items that were no longer referenced.

To start with I did the obvious thing and made the allocation region larger, ignoring the problem.  But eventually that too became untenable.  So now I've implemented a stop & copy garbage collector, using Cheney's algorithm.   Every time our `(cons ..)` primitive is called we check to see if we should invoke the garbage-collection process.

The `(cons ..)` primitive is a lisp-fundamental, so I figure that is going to be called pretty often in user-programs, either directly or via the `(list ...)` wrapper.  But if that isn't the case you may also trigger the garbage-collection process explicitly, and see the stats via these methods:

* `(sys-gc)` - run the garbage collection process immediately.
* `(sys-gc-threshold)` - We run the GC process after a fixed number of allocations have been made, this retrieves the value of that threshold.  Change it via the `GC_ALLOC_THRESHOLD` environmental variable.
* `(sys-heap-allocs)` -  Return the number of memory allocations made.
* `(sys-heap-bytes)` - Return the size of the heap.
* `(sys-heap-data)` - Return the contents of the heap as a list of entries.
* `(sys-heap-dump)` - Dump a summary of the heap to the console.
  * This is implemented in assembly and literally writes to STDOUT.  There is no control over the formatting.
  * Use `sys-heap-data` if you want to get the heap-data and format it for printing yourself.
* `(sys-heap-objects)` - Return the number of objects stored upon the heap.

The stop and copy implementation is pretty simple:

* We have *two* heap areas, each of which are an identical size.
* One heap is used as the backing-store for all allocations we make.
* When a `sys-gc` request is made the current heap is inspected and all live items are copied to the other heap.
  * The new heap is then made the active one, which essentially orphans and frees the unreachable entries upon the old heap.
  * As a nice side-effect this removes any fragmentation, there are no holes in the new heap.  Allocation continues to grow the heap with no need to worry about using previously-freed objects.
* The copying process has to deal with global variables, objects held within stack-frames, and those objects which might be held inside registers.
  * For register contents we cheat a little.
  * The `(cons ..)` primitive is the only one that is used to trigger "auto GC",  and we know `cons` can only be called with two arguments, so we only have to consider the two registers RDI & RSI.
  * TLDR; Our roots are "globals", "stack-locals", and potentially the contents of the two registers `rdi` and `rsi`.

By default the garbage collection process is triggered when more than 1000 allocations have been made, but if you wish you can adjust the threshold by setting `GC_ALLOC_THRESHOLD` environmental variable before you execute one of the compiled binaries:

    GC_ALLOC_THRESHOLD=1      ./foo
    GC_ALLOC_THRESHOLD=500000 ./foo

Setting the threshold low will ensure the heap is as small as possible, at a cost that the garbage collector will run more frequently.  Setting it to a "mid-high" value is perhaps more efficient, the garbage collector will run less often, and do more work each time.  But chances are that time won't dominate the _total_ runtime, which is a risk if it is set too low.

The function `at_exit` is invoked, if it is defined, when all programs terminate cleanly.  The default handler will dump memory statistics on-exit if the environmental variable `GC_DUMP_STATS` is non-empty.
