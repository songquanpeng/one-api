import { useState, useEffect } from "react";
import { useSelector } from "react-redux";
import Turnstile from "react-turnstile";
import { API } from "utils/api";

// material-ui
import { useTheme } from "@mui/material/styles";
import {
  Box,
  Button,
  FormControl,
  FormHelperText,
  InputLabel,
  OutlinedInput,
  Typography,
} from "@mui/material";

// third party
import * as Yup from "yup";
import { Formik } from "formik";

// project imports
import AnimateButton from "ui-component/extended/AnimateButton";

// assets
import { showError, showInfo, showSuccess } from "utils/common";

// ===========================|| FIREBASE - REGISTER ||=========================== //

const ForgetPasswordForm = ({ ...others }) => {
  const theme = useTheme();
  const siteInfo = useSelector((state) => state.siteInfo);

  const [sendEmail, setSendEmail] = useState(false);
  const [turnstileEnabled, setTurnstileEnabled] = useState(false);
  const [turnstileSiteKey, setTurnstileSiteKey] = useState("");
  const [turnstileToken, setTurnstileToken] = useState("");
  const [disableButton, setDisableButton] = useState(false);
  const [countdown, setCountdown] = useState(30);

  const submit = async (values, { setSubmitting }) => {
    setDisableButton(true);
    setSubmitting(true);
    if (turnstileEnabled && turnstileToken === "") {
      showInfo("请稍后几秒重试，Turnstile 正在检查用户环境！");
      setSubmitting(false);
      return;
    }
    const res = await API.get(
      `/api/reset_password?email=${values.email}&turnstile=${turnstileToken}`
    );
    const { success, message } = res.data;
    if (success) {
      showSuccess("重置邮件发送成功，请检查邮箱！");
      setSendEmail(true);
    } else {
      showError(message);
      setDisableButton(false);
      setCountdown(30);
    }
    setSubmitting(false);
  };

  useEffect(() => {
    let countdownInterval = null;
    if (disableButton && countdown > 0) {
      countdownInterval = setInterval(() => {
        setCountdown(countdown - 1);
      }, 1000);
    } else if (countdown === 0) {
      setDisableButton(false);
      setCountdown(30);
    }
    return () => clearInterval(countdownInterval);
  }, [disableButton, countdown]);

  useEffect(() => {
    if (siteInfo.turnstile_check) {
      setTurnstileEnabled(true);
      setTurnstileSiteKey(siteInfo.turnstile_site_key);
    }
  }, [siteInfo]);

  return (
    <>
      {sendEmail ? (
        <Typography variant="h3" padding={"20px"}>
          重置邮件发送成功，请检查邮箱！
        </Typography>
      ) : (
        <Formik
          initialValues={{
            email: "",
          }}
          validationSchema={Yup.object().shape({
            email: Yup.string()
              .email("必须是有效的Email地址")
              .max(255)
              .required("Email是必填项"),
          })}
          onSubmit={submit}
        >
          {({
            errors,
            handleBlur,
            handleChange,
            handleSubmit,
            isSubmitting,
            touched,
            values,
          }) => (
            <form noValidate onSubmit={handleSubmit} {...others}>
              <FormControl
                fullWidth
                error={Boolean(touched.email && errors.email)}
                sx={{ ...theme.typography.customInput }}
              >
                <InputLabel htmlFor="outlined-adornment-email-register">
                  Email
                </InputLabel>
                <OutlinedInput
                  id="outlined-adornment-email-register"
                  type="text"
                  value={values.email}
                  name="email"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  inputProps={{}}
                />
                {touched.email && errors.email && (
                  <FormHelperText
                    error
                    id="standard-weight-helper-text--register"
                  >
                    {errors.email}
                  </FormHelperText>
                )}
              </FormControl>

              {turnstileEnabled ? (
                <Turnstile
                  sitekey={turnstileSiteKey}
                  onVerify={(token) => {
                    setTurnstileToken(token);
                  }}
                />
              ) : (
                <></>
              )}

              <Box sx={{ mt: 2 }}>
                <AnimateButton>
                  <Button
                    disableElevation
                    disabled={isSubmitting || disableButton}
                    fullWidth
                    size="large"
                    type="submit"
                    variant="contained"
                    color="primary"
                  >
                    {disableButton ? `重试 (${countdown})` : "提交"}
                  </Button>
                </AnimateButton>
              </Box>
            </form>
          )}
        </Formik>
      )}
    </>
  );
};

export default ForgetPasswordForm;
