/*
 * UserCard.js
 *
 * This file uses code from the Minimal UI project, available at
 * https://github.com/minimal-ui-kit/material-kit-react/blob/main/src/sections/blog/post-card.jsx
 *
 * Minimal UI is licensed under the MIT License. A copy of the license is included below:
 *
 * MIT License
 *
 * Copyright (c) 2021 Minimal UI (https://minimals.cc/)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */
import { Box, Avatar } from '@mui/material';
import { alpha } from '@mui/material/styles';
import Card from '@mui/material/Card';
import shapeAvatar from 'assets/images/icons/shape-avatar.svg';
import coverAvatar from 'assets/images/invite/cover.jpg';
import userAvatar from 'assets/images/users/user-round.svg';
import SvgColor from 'ui-component/SvgColor';

import React from 'react';

export default function UserCard({ children }) {
  const renderShape = (
    <SvgColor
      color="paper"
      src={shapeAvatar}
      sx={{
        width: '100%',
        height: 62,
        zIndex: 10,
        bottom: -26,
        position: 'absolute',
        color: 'background.paper'
      }}
    />
  );

  const renderAvatar = (
    <Avatar
      src={userAvatar}
      sx={{
        zIndex: 11,
        width: 64,
        height: 64,
        position: 'absolute',
        alignItems: 'center',
        marginLeft: 'auto',
        marginRight: 'auto',
        left: 0,
        right: 0,
        bottom: (theme) => theme.spacing(-4)
      }}
    />
  );

  const renderCover = (
    <Box
      component="img"
      src={coverAvatar}
      sx={{
        top: 0,
        width: 1,
        height: 1,
        objectFit: 'cover',
        position: 'absolute'
      }}
    />
  );

  return (
    <Card>
      <Box
        sx={{
          position: 'relative',
          '&:after': {
            top: 0,
            content: "''",
            width: '100%',
            height: '100%',
            position: 'absolute',
            bgcolor: (theme) => alpha(theme.palette.primary.main, 0.42)
          },
          pt: {
            xs: 'calc(100% / 3)',
            sm: 'calc(100% / 4.66)'
          }
        }}
      >
        {renderShape}
        {renderAvatar}
        {renderCover}
      </Box>
      <Box
        sx={{
          p: (theme) => theme.spacing(4, 3, 3, 3)
        }}
      >
        {children}
      </Box>
    </Card>
  );
}
