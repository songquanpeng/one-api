/*
 * Label.js
 *
 * This file uses code from the Minimal UI project, available at
 * https://github.com/minimal-ui-kit/material-kit-react/blob/main/src/components/label/label.jsx
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
import PropTypes from 'prop-types';
import { forwardRef } from 'react';

import Box from '@mui/material/Box';
import { useTheme } from '@mui/material/styles';
import { alpha, styled } from '@mui/material/styles';

// ----------------------------------------------------------------------

const Label = forwardRef(({ children, color = 'default', variant = 'soft', startIcon, endIcon, sx, ...other }, ref) => {
  const theme = useTheme();

  const iconStyles = {
    width: 16,
    height: 16,
    '& svg, img': { width: 1, height: 1, objectFit: 'cover' }
  };

  return (
    <StyledLabel
      ref={ref}
      component="span"
      ownerState={{ color, variant }}
      sx={{
        ...(startIcon && { pl: 0.75 }),
        ...(endIcon && { pr: 0.75 }),
        ...sx
      }}
      theme={theme}
      {...other}
    >
      {startIcon && <Box sx={{ mr: 0.75, ...iconStyles }}> {startIcon} </Box>}

      {children}

      {endIcon && <Box sx={{ ml: 0.75, ...iconStyles }}> {endIcon} </Box>}
    </StyledLabel>
  );
});

Label.propTypes = {
  children: PropTypes.node,
  endIcon: PropTypes.object,
  startIcon: PropTypes.object,
  sx: PropTypes.object,
  variant: PropTypes.oneOf(['filled', 'outlined', 'ghost', 'soft']),
  color: PropTypes.oneOf(['default', 'primary', 'secondary', 'info', 'success', 'warning', 'orange', 'error'])
};

export default Label;

const StyledLabel = styled(Box)(({ theme, ownerState }) => {
  // const lightMode = theme.palette.mode === 'light';

  const filledVariant = ownerState.variant === 'filled';

  const outlinedVariant = ownerState.variant === 'outlined';

  const softVariant = ownerState.variant === 'soft';

  const ghostVariant = ownerState.variant === 'ghost';

  const defaultStyle = {
    ...(ownerState.color === 'default' && {
      // FILLED
      ...(filledVariant && {
        color: theme.palette.grey[300],
        backgroundColor: theme.palette.text.primary
      }),
      // OUTLINED
      ...(outlinedVariant && {
        color: theme.palette.grey[500],
        border: `2px solid ${theme.palette.grey[500]}`
      }),
      // SOFT
      ...(softVariant && {
        color: theme.palette.text.secondary,
        backgroundColor: alpha(theme.palette.grey[500], 0.16)
      })
    })
  };

  const colorStyle = {
    ...(ownerState.color !== 'default' && {
      // FILLED
      ...(filledVariant && {
        color: theme.palette.background.paper,
        backgroundColor: theme.palette[ownerState.color]?.main
      }),
      // OUTLINED
      ...(outlinedVariant && {
        backgroundColor: 'transparent',
        color: theme.palette[ownerState.color]?.main,
        border: `2px solid ${theme.palette[ownerState.color]?.main}`
      }),
      // SOFT
      ...(softVariant && {
        color: theme.palette[ownerState.color]['dark'],
        backgroundColor: alpha(theme.palette[ownerState.color]?.main, 0.16)
      }),
      // GHOST
      ...(ghostVariant && {
        color: theme.palette[ownerState.color]?.main
      })
    })
  };

  return {
    height: 24,
    minWidth: 24,
    lineHeight: 0,
    borderRadius: 6,
    cursor: 'default',
    alignItems: 'center',
    whiteSpace: 'nowrap',
    display: 'inline-flex',
    justifyContent: 'center',
    // textTransform: 'capitalize',
    padding: theme.spacing(0, 0.75),
    fontSize: theme.typography.pxToRem(12),
    fontWeight: theme.typography.fontWeightBold,
    transition: theme.transitions.create('all', {
      duration: theme.transitions.duration.shorter
    }),
    ...defaultStyle,
    ...colorStyle
  };
});
