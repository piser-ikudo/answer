/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import qs from 'qs';

import request from '@/utils/request';
import type * as Type from '@/common/interface';

/**
 * axios serializes arrays as `tag_groups[]=a&tag_groups[]=b` by default, but the
 * API binds the plain name, so the value would be dropped and the tag group
 * filter would silently do nothing. `arrayFormat: 'repeat'` sends
 * `tag_groups=a&tag_groups=b` instead, which is what the API expects.
 */
const paramsSerializer = (params: unknown) =>
  qs.stringify(params, { arrayFormat: 'repeat' });

export const getSearchResult = (params?: Type.SearchParams) => {
  const apiUrl = '/answer/api/v1/search';

  return request.get<Type.SearchRes>(apiUrl, {
    params,
    paramsSerializer,
  });
};
