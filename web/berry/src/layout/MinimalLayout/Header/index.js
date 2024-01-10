// material-ui
import { useTheme } from "@mui/material/styles";
import { Box, Button, Stack } from "@mui/material";
import LogoSection from "layout/MainLayout/LogoSection";
import { Link } from "react-router-dom";
import { useLocation } from "react-router-dom";
import { useSelector } from "react-redux";

// ==============================|| MAIN NAVBAR / HEADER ||============================== //

const Header = () => {
  const theme = useTheme();
  const { pathname } = useLocation();
  const account = useSelector((state) => state.account);

  return (
    <>
      <Box
        sx={{
          width: 228,
          display: "flex",
          [theme.breakpoints.down("md")]: {
            width: "auto",
          },
        }}
      >
        <Box component="span" sx={{ flexGrow: 1 }}>
          <LogoSection />
        </Box>
      </Box>

      <Box sx={{ flexGrow: 1 }} />
      <Box sx={{ flexGrow: 1 }} />
      <Stack spacing={2} direction="row">
        <Button
          component={Link}
          variant="text"
          to="/"
          color={pathname === "/" ? "primary" : "inherit"}
        >
          首页
        </Button>
        <Button
          component={Link}
          variant="text"
          to="/about"
          color={pathname === "/about" ? "primary" : "inherit"}
        >
          关于
        </Button>
        {account.user ? (
          <Button
            component={Link}
            variant="contained"
            to="/panel"
            color="primary"
          >
            控制台
          </Button>
        ) : (
          <Button
            component={Link}
            variant="contained"
            to="/login"
            color="primary"
          >
            登入
          </Button>
        )}
      </Stack>
    </>
  );
};

export default Header;
