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

import React, { useEffect, useState } from 'react';
import { Button } from 'react-bootstrap';
import { useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { useQueryTagGroups } from '@/services';

/**
 * Tag group multi select for the advanced search card.
 *
 * Every tag group of the site is listed. The selection is applied to the
 * `tag_groups` url parameter only when the button is pressed, so ticking several
 * boxes runs a single search. The search box may stay empty, then every post
 * whose tags belong to the selected groups is returned.
 */
const CheckboxGroup: React.FC = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'search.tips' });
  const [urlSearchParams, setUrlSearchParams] = useSearchParams();
  const { data: tagGroups } = useQueryTagGroups();

  const groups = tagGroups?.groups || [];
  const appliedRaw = urlSearchParams.get('tag_groups') || '';
  // keep the ticks editable without triggering a request on every click
  const [picked, setPicked] = useState<string[]>(
    appliedRaw.split(',').filter(Boolean),
  );

  // follow the url when it changes from outside, e.g. browser navigation
  useEffect(() => {
    setPicked(appliedRaw.split(',').filter(Boolean));
  }, [appliedRaw]);

  const handleCheckboxChange = (tagGroup: string) => {
    setPicked((prev) =>
      prev.includes(tagGroup)
        ? prev.filter((item) => item !== tagGroup)
        : [...prev, tagGroup],
    );
  };

  const handleSearch = () => {
    const next = new URLSearchParams(urlSearchParams);
    if (picked.length > 0) {
      next.set('tag_groups', picked.join(','));
    } else {
      next.delete('tag_groups');
    }
    // a new filter always starts from the first page
    next.set('page', '1');
    setUrlSearchParams(next);
  };

  const handleClear = () => {
    setPicked([]);
    const next = new URLSearchParams(urlSearchParams);
    next.delete('tag_groups');
    next.set('page', '1');
    setUrlSearchParams(next);
  };

  return (
    <div className="checkbox-group p-3">
      {groups.length === 0 ? (
        <div className="text-secondary small">{t('tag_groups_empty')}</div>
      ) : (
        <>
          <div className="text-secondary small mb-2">{t('tag_groups')}</div>
          {groups.map((group) => (
            <div key={group.tag_group} className="form-check">
              <input
                type="checkbox"
                className="form-check-input"
                id={`tag-group-${group.tag_group || 'ungrouped'}`}
                checked={picked.includes(group.tag_group)}
                onChange={() => handleCheckboxChange(group.tag_group)}
              />
              <label
                className="form-check-label"
                htmlFor={`tag-group-${group.tag_group || 'ungrouped'}`}>
                {group.tag_group}
                <span className="text-secondary small ms-1">
                  ({group.tag_count})
                </span>
              </label>
            </div>
          ))}
          <div className="d-flex align-items-center gap-2 mt-3">
            <Button size="sm" variant="primary" onClick={handleSearch}>
              {t('tag_groups_btn')}
            </Button>
            {picked.length > 0 && (
              <Button size="sm" variant="link" onClick={handleClear}>
                {t('tag_groups_clear')}
              </Button>
            )}
          </div>
        </>
      )}
    </div>
  );
};

export default CheckboxGroup;
