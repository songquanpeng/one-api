// material-ui
import Skeleton from '@mui/material/Skeleton';

// ==============================|| SKELETON IMAGE CARD ||============================== //

const ImagePlaceholder = ({ ...others }) => <Skeleton variant="rectangular" {...others} animation="wave" />;

export default ImagePlaceholder;
