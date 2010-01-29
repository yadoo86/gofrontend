/* go-refcount-increment.c -- increment reference count.

   Copyright 2009 The Go Authors. All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.  */

#include <assert.h>
#include <stdlib.h>

#include "go-refcount.h"

/* Increment the reference count for a pointer.  */

void
__go_increment_refcount (void *value,
			 const struct __go_type_descriptor *descriptor)
{
  struct
  {
    struct __go_refcount q;
    struct __go_refcount_entry e;
  } s;
  struct __go_refcount *queue;

  queue = &s.q;
  __builtin_memset (queue, 0, sizeof (struct __go_refcount));
  queue->__caller = __go_refcount_head;
  __go_refcount_head = queue;
  queue->__increment_copy_count = 1;
  queue->__entries[0].__descriptor = descriptor;
  queue->__entries[0].__value = value;
  __go_refcount_flush_queue (queue);
  __go_refcount_head = queue->__caller;
}
