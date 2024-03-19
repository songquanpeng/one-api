import PropTypes from "prop-types";
import Label from "ui-component/Label";
import Stack from "@mui/material/Stack";
import Divider from "@mui/material/Divider";

function name2color(name) {
  switch (name) {
    case "default":
      return "info";
    case "vip":
      return "warning"
    case "svip":
      return "error"
    default:
      return "info"
  }
}

const GroupLabel = ({ group }) => {
  let groups = [];
  if (group === "") {
    groups = ["default"];
  } else {
    groups = group.split(",");
    groups.sort();
  }
  return (
    <Stack divider={<Divider orientation="vertical" flexItem />} spacing={0.5}>
      {groups.map((group, index) => {
        return <Label key={index} color={name2color(group)}>{group}</Label>;
      })}
    </Stack>
  );
};

GroupLabel.propTypes = {
  group: PropTypes.string,
};

export default GroupLabel;
