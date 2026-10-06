import { z } from '#/adapter/form';
import { $t } from '#/locales';

/**
 * 前端口令规则，与后端 base-kit 口令策略的结构部分一致：8~64 位、同时含字母和数字。
 * 常见弱口令、包含用户名、系统配置 password_min_length 调高的长度由后端判断，不合格时接口返回原因。
 */
export const PASSWORD_MIN_LENGTH = 8;
export const PASSWORD_MAX_LENGTH = 64;

function passwordIssue(value: string, label: string) {
  if (value.length < PASSWORD_MIN_LENGTH) {
    return $t('ui.formRules.minLength', [label, PASSWORD_MIN_LENGTH]);
  }
  if (value.length > PASSWORD_MAX_LENGTH) {
    return $t('ui.formRules.maxLength', [label, PASSWORD_MAX_LENGTH]);
  }
  if (!/\p{L}/u.test(value) || !/\p{N}/u.test(value)) {
    return $t('system.password.letterAndDigit');
  }
  return undefined;
}

/**
 * 口令字段的校验规则。optional 为 true 时留空合法（编辑用户时「留空则不修改」）。
 */
export function passwordRule(label: string, optional = false) {
  const required = $t('ui.formRules.required', [label]);
  const rule = z
    .string({ required_error: required })
    .superRefine((value, ctx) => {
      if (!value) {
        if (!optional)
          ctx.addIssue({ code: z.ZodIssueCode.custom, message: required });
        return;
      }
      const issue = passwordIssue(value, label);
      if (issue) ctx.addIssue({ code: z.ZodIssueCode.custom, message: issue });
    });
  return optional ? rule.optional() : rule;
}
