import { useState, useEffect } from 'react';
import SubCard from 'ui-component/cards/SubCard';
// import { gridSpacing } from 'store/constant';
import { API } from 'utils/api';
import { showError, copy } from 'utils/common';
import { Typography, Accordion, AccordionSummary, AccordionDetails, Box, Stack } from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import Label from 'ui-component/Label';

const SupportModels = () => {
  const [modelList, setModelList] = useState([]);

  const fetchModels = async () => {
    try {
      let res = await API.get(`/api/user/models`);
      if (res === undefined) {
        return;
      }
      // 对 res.data.data 里面的 owned_by 进行分组
      let modelGroup = {};
      res.data.data.forEach((model) => {
        if (modelGroup[model.owned_by] === undefined) {
          modelGroup[model.owned_by] = [];
        }
        modelGroup[model.owned_by].push(model.id);
      });
      setModelList(modelGroup);
    } catch (error) {
      showError(error.message);
    }
  };

  useEffect(() => {
    fetchModels();
  }, []);

  return (
    <Accordion key="support_models" sx={{ borderRadius: '12px' }}>
      <AccordionSummary aria-controls="support_models" expandIcon={<ExpandMoreIcon />}>
        <Typography variant="subtitle1">当前可用模型</Typography>
      </AccordionSummary>
      <AccordionDetails>
        <Stack spacing={1}>
          {Object.entries(modelList).map(([title, models]) => (
            <SubCard key={title} title={title === 'null' ? '其他模型' : title}>
              <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: '10px' }}>
                {models.map((model) => (
                  <Label
                    variant="outlined"
                    color="primary"
                    key={model}
                    onClick={() => {
                      copy(model, '模型名称');
                    }}
                  >
                    {model}
                  </Label>
                ))}
              </Box>
            </SubCard>
          ))}
        </Stack>
      </AccordionDetails>
    </Accordion>
  );
};

export default SupportModels;
