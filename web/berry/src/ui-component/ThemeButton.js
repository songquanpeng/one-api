import { useDispatch, useSelector } from 'react-redux';
import { SET_THEME } from 'store/actions';
import { useTheme } from '@mui/material/styles';
import { Avatar, Box, ButtonBase } from '@mui/material';
import { IconSun, IconMoon } from '@tabler/icons-react';

export default function ThemeButton() {
  const dispatch = useDispatch();

  const defaultTheme = useSelector((state) => state.customization.theme);

  const theme = useTheme();

  return (
    <Box
      sx={{
        ml: 2,
        mr: 3,
        [theme.breakpoints.down('md')]: {
          mr: 2
        }
      }}
    >
      <ButtonBase sx={{ borderRadius: '12px' }}>
        <Avatar
          variant="rounded"
          sx={{
            ...theme.typography.commonAvatar,
            ...theme.typography.mediumAvatar,
            transition: 'all .2s ease-in-out',
            borderColor: theme.typography.menuChip.background,
            backgroundColor: theme.typography.menuChip.background,
            '&[aria-controls="menu-list-grow"],&:hover': {
              background: theme.palette.secondary.dark,
              color: theme.palette.secondary.light
            }
          }}
          onClick={() => {
            let theme = defaultTheme === 'light' ? 'dark' : 'light';
            dispatch({ type: SET_THEME, theme: theme });
            localStorage.setItem('theme', theme);
          }}
          color="inherit"
        >
          {defaultTheme === 'light' ? <IconSun stroke={1.5} size="1.3rem" /> : <IconMoon stroke={1.5} size="1.3rem" />}
        </Avatar>
      </ButtonBase>
    </Box>
  );
}
