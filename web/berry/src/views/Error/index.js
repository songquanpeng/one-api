import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Container from '@mui/material/Container';
import NotFound from 'assets/images/404.svg';
import { useNavigate } from 'react-router';

// ----------------------------------------------------------------------

export default function NotFoundView() {
  const navigate = useNavigate();
  const goBack = () => {
    navigate(-1);
  };
  return (
    <>
      <Container>
        <Box
          sx={{
            py: 12,
            maxWidth: 480,
            mx: 'auto',
            display: 'flex',
            minHeight: 'calc(100vh - 136px)',
            textAlign: 'center',
            alignItems: 'center',
            flexDirection: 'column',
            justifyContent: 'center'
          }}
        >
          <Box
            component="img"
            src={NotFound}
            sx={{
              mx: 'auto',
              height: 260,
              my: { xs: 5, sm: 10 }
            }}
          />

          <Button size="large" variant="contained" onClick={goBack}>
            返回
          </Button>
        </Box>
      </Container>
    </>
  );
}
