import PropTypes from 'prop-types';
import SubCard from 'ui-component/cards/SubCard';
import { Typography, Tooltip, Divider } from '@mui/material';
import SkeletonDataCard from 'ui-component/cards/Skeleton/DataCard';

export default function DataCard({ isLoading, title, content, tip, subContent }) {
  return (
    <>
      {isLoading ? (
        <SkeletonDataCard />
      ) : (
        <SubCard sx={{ height: '160px' }}>
          <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
            {title}
          </Typography>
          <Typography variant="h3" sx={{ fontSize: '2rem', lineHeight: 1.5, fontWeight: 700 }}>
            {tip ? (
              <Tooltip title={tip} placement="top">
                <span>{content}</span>
              </Tooltip>
            ) : (
              content
            )}
          </Typography>
          <Divider />
          <Typography variant="subtitle2" sx={{ mt: 2 }}>
            {subContent}
          </Typography>
        </SubCard>
      )}
    </>
  );
}

DataCard.propTypes = {
  isLoading: PropTypes.bool,
  title: PropTypes.string,
  content: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
  tip: PropTypes.node,
  subContent: PropTypes.node
};
