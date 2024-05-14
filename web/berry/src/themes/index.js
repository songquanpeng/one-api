import { createTheme } from '@mui/material/styles';

// assets
import colors from 'assets/scss/_themes-vars.module.scss';

// project imports
import componentStyleOverrides from './compStyleOverride';
import themePalette from './palette';
import themeTypography from './typography';

/**
 * Represent theme style and structure as per Material-UI
 * @param {JsonObject} customization customization parameter object
 */

export const theme = (customization) => {
  const color = colors;
  const options = customization.theme === 'light' ? GetLightOption() : GetDarkOption();
  const themeOption = {
    colors: color,
    ...options,
    customization
  };

  const themeOptions = {
    direction: 'ltr',
    palette: themePalette(themeOption),
    mixins: {
      toolbar: {
        minHeight: '48px',
        padding: '16px',
        '@media (min-width: 600px)': {
          minHeight: '48px'
        }
      }
    },
    typography: themeTypography(themeOption)
  };

  const themes = createTheme(themeOptions);
  themes.components = componentStyleOverrides(themeOption);

  return themes;
};

export default theme;

function GetDarkOption() {
  const color = colors;
  return {
    mode: 'dark',
    heading: color.darkTextTitle,
    paper: color.darkLevel2,
    backgroundDefault: color.darkPaper,
    background: color.darkBackground,
    darkTextPrimary: color.darkTextPrimary,
    darkTextSecondary: color.darkTextSecondary,
    textDark: color.darkTextTitle,
    menuSelected: color.darkSecondaryMain,
    menuSelectedBack: color.darkSelectedBack,
    divider: color.darkDivider,
    borderColor: color.darkBorderColor,
    menuButton: color.darkLevel1,
    menuButtonColor: color.darkSecondaryMain,
    menuChip: color.darkLevel1,
    headBackgroundColor: color.darkBackground,
    tableBorderBottom: color.darkDivider
  };
}

function GetLightOption() {
  const color = colors;
  return {
    mode: 'light',
    heading: color.grey900,
    paper: color.paper,
    backgroundDefault: color.paper,
    background: color.primaryLight,
    darkTextPrimary: color.grey700,
    darkTextSecondary: color.grey500,
    textDark: color.grey900,
    menuSelected: color.secondaryDark,
    menuSelectedBack: color.secondaryLight,
    divider: color.grey200,
    borderColor: color.grey300,
    menuButton: color.secondaryLight,
    menuButtonColor: color.secondaryDark,
    menuChip: color.primaryLight,
    headBackgroundColor: color.tableBackground,
    tableBorderBottom: color.tableBorderBottom
  };
}
