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

import { FC, useEffect, useState } from 'react';
import { Dropdown, Form } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { useQueryTagGroups } from '@/services';
import { formatCount } from '@/utils';
import './index.scss';

interface Props {
  className?: string;
  /** currently selected tag groups */
  selected?: string[];
  /** called with the whole selection whenever a group is toggled */
  onChange: (groups: string[]) => void;
  size?: 'sm' | 'lg';
}

const TagGroupFilter: FC<Props> = ({
  className = '',
  selected = [],
  onChange,
  size,
}) => {
  const { t } = useTranslation('translation', { keyPrefix: 'tags' });
  const { data: tagGroups } = useQueryTagGroups();
  // keep a local selection so the checkbox state is correct even before the
  // parent has propagated the new value back down
  const [picked, setPicked] = useState<string[]>(selected);

  useEffect(() => {
    setPicked(selected);
  }, [selected]);

  const groups = tagGroups?.groups || [];

  const buildLabel = () => {
    if (picked.length === 0) {
      return t('group_filter_all');
    }
    if (picked.length === 1) {
      return picked[0] || t('group_filter_ungrouped');
    }
    return t('group_filter_selected', { count: picked.length });
  };

  const handleToggle = (group: string) => {
    const next = picked.includes(group)
      ? picked.filter((item) => item !== group)
      : [...picked, group];
    setPicked(next);
    onChange(next);
  };

  const handleClear = () => {
    setPicked([]);
    onChange([]);
  };

  return (
    <Dropdown
      autoClose="outside"
      className={`tag-group-filter ${className}`}
      id="tag-group-filter">
      <Dropdown.Toggle variant="outline-secondary" size={size}>
        {buildLabel()}
      </Dropdown.Toggle>
      <Dropdown.Menu className="tag-group-filter-menu">
        {groups.length === 0 ? (
          <Dropdown.Header>{t('group_filter_empty')}</Dropdown.Header>
        ) : (
          <>
            {picked.length > 0 && (
              <Dropdown.Item as="button" type="button" onClick={handleClear}>
                {t('group_filter_clear')}
              </Dropdown.Item>
            )}
            {groups.map((group) => (
              <Dropdown.Item
                key={group.tag_group}
                as="div"
                className="tag-group-filter-item">
                <Form.Check
                  type="checkbox"
                  id={`tag-group-${group.tag_group || 'ungrouped'}`}
                  checked={picked.includes(group.tag_group)}
                  onChange={() => handleToggle(group.tag_group)}
                  label={`${
                    group.tag_group || t('group_filter_ungrouped')
                  } (${formatCount(group.tag_count)})`}
                />
              </Dropdown.Item>
            ))}
          </>
        )}
      </Dropdown.Menu>
    </Dropdown>
  );
};

export default TagGroupFilter;
