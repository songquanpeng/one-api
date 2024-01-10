import { Box, Typography, Button, Container, Stack } from '@mui/material';
import Grid from '@mui/material/Unstable_Grid2';
import { GitHub } from '@mui/icons-material';

const BaseIndex = () => (
  <Box
    sx={{
      minHeight: 'calc(100vh - 136px)',
      backgroundImage: 'linear-gradient(to right, #ff9966, #ff5e62)',
      color: 'white',
      p: 4
    }}
  >
    <Container maxWidth="lg">
      <Grid container columns={12} wrap="nowrap" alignItems="center" sx={{ minHeight: 'calc(100vh - 230px)' }}>
        <Grid md={7} lg={6}>
          <Stack spacing={3}>
            <Typography variant="h1" sx={{ fontSize: '4rem', color: '#fff', lineHeight: 1.5 }}>
              One API
            </Typography>
            <Typography variant="h4" sx={{ fontSize: '1.5rem', color: '#fff', lineHeight: 1.5 }}>
              All in one 的 OpenAI 接口 <br />
              整合各种 API 访问方式 <br />
              一键部署，开箱即用
            </Typography>
            <Button
              variant="contained"
              startIcon={<GitHub />}
              href="https://github.com/songquanpeng/one-api"
              target="_blank"
              sx={{ backgroundColor: '#24292e', color: '#fff', width: 'fit-content', boxShadow: '0 3px 5px 2px rgba(255, 105, 135, .3)' }}
            >
              GitHub
            </Button>
          </Stack>
        </Grid>
      </Grid>
    </Container>
  </Box>
);

export default BaseIndex;
