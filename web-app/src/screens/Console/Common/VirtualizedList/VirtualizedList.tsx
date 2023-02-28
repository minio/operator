// This file is part of MinIO Operator
// Copyright (c) 2022 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

import React, { Fragment, ReactElement } from "react";
import { FixedSizeList as List } from "react-window";
import InfiniteLoader from "react-window-infinite-loader";
import { AutoSizer } from "react-virtualized";

interface IVirtualizedList {
  rowRenderFunction: (index: number) => ReactElement | null;
  totalItems: number;
  defaultHeight?: number;
}

let itemStatusMap: any = {};
const LOADING = 1;
const LOADED = 2;

const VirtualizedList = ({
  rowRenderFunction,
  totalItems,
  defaultHeight,
}: IVirtualizedList) => {
  const isItemLoaded = (index: any) => !!itemStatusMap[index];

  const loadMoreItems = (startIndex: number, stopIndex: number) => {
    for (let index = startIndex; index <= stopIndex; index++) {
      itemStatusMap[index] = LOADING;
    }

    for (let index = startIndex; index <= stopIndex; index++) {
      itemStatusMap[index] = LOADED;
    }
  };

  const RenderItemLine = ({ index, style }: any) => {
    return <div style={style}>{rowRenderFunction(index)}</div>;
  };

  return (
    <Fragment>
      <InfiniteLoader
        isItemLoaded={isItemLoaded}
        loadMoreItems={loadMoreItems}
        itemCount={totalItems}
      >
        {({ onItemsRendered, ref }) => (
          // @ts-ignore
          <AutoSizer>
            {({ width, height }) => {
              return (
                <List
                  itemSize={defaultHeight || 220}
                  height={height}
                  itemCount={totalItems}
                  width={width}
                  ref={ref}
                  onItemsRendered={onItemsRendered}
                >
                  {RenderItemLine}
                </List>
              );
            }}
          </AutoSizer>
        )}
      </InfiniteLoader>
    </Fragment>
  );
};

export default VirtualizedList;
