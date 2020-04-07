/*
==================================================================================
  Copyright (c) 2019 AT&T Intellectual Property.
  Copyright (c) 2019 Nokia

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
==================================================================================
*/

#ifndef __MEMTRACK_H__
#define __MEMTRACK_H__


#ifdef __cplusplus
extern "C" {
#endif
	
/*
 *
 */
typedef struct mem_track_s{
    struct mem_track_s *next;
    size_t  sz;
    unsigned char ptr[];
}mem_track_t;


/*
 *
 */
typedef struct mem_track_hdr_s{
    struct mem_track_s *next;
}mem_track_hdr_t;

#define MEM_TRACK_HDR_INIT {0}


/*
 *
 */
void mem_track_init(mem_track_hdr_t *curr);

void* mem_track_alloc(mem_track_hdr_t *curr, size_t sz);

void mem_track_free(mem_track_hdr_t *curr);


/*
 *
 */
#define MEM_TRACK_ALLOC(__hdr,__id) (__id*)mem_track_alloc(__hdr,sizeof(__id))

#define MEM_TRACK_ALLOC_LIST(__hdr,__id,__n) (__id**)mem_track_alloc(__hdr,sizeof(__id)*__n)

#define MEM_TRACK_ALLOC_PTR_LIST(__hdr,__id,__n) (__id**)mem_track_alloc(__hdr,sizeof(__id*)*__n)

#define MEM_TRACK_ALLOC_BUFFER(__hdr,__id,__n) (__id*)mem_track_alloc(__hdr,sizeof(__id)*__n)

#define MEM_TRACK_ALLOC_INIT(__val, __hdr,__id,__init) __id* __val=MEM_TRACK_ALLOC(__hdr,__id); *__val=(__id)__init

#ifdef __cplusplus
}
#endif

#endif
