import { baseRequestClient, requestClient } from '#/api/request';

export namespace AuthApi {
  /** 登录接口参数 */
  export interface LoginParams {
    /** 系统配置 login_captcha 开启时必填：拼图验证通过后拿到的一次性凭证 */
    captchaToken?: string;
    password?: string;
    username?: string;
  }

  /** 登录接口返回值 */
  export interface LoginResult {
    accessToken: string;
  }

  /** 登录拼图验证码；未开启时只有 enabled=false。坐标、尺寸都按背景图原始像素 */
  export interface CaptchaResult {
    captchaId?: string;
    enabled: boolean;
    height?: number;
    /** 带缺口的背景图（data URI） */
    image?: string;
    /** 拼图块（data URI，透明底） */
    piece?: string;
    pieceSize?: number;
    /** 拼图块顶边的纵坐标；横向从 0 开始拖 */
    pieceY?: number;
    width?: number;
  }

  /** 提交拼图位置 */
  export interface CaptchaVerifyParams {
    captchaId: string;
    /** 拼图块左边缘拖到的横坐标（背景图原始像素） */
    x: number;
  }

  /** 拼图验证接口的原始响应（走 baseRequestClient，不经统一拦截器） */
  export interface CaptchaVerifyRawResult {
    data: { data?: { token: string }; message?: string };
  }

  export interface RefreshTokenResult {
    data: string;
    status: number;
  }

  /** 修改当前用户密码参数 */
  export interface ChangePasswordParams {
    newPassword: string;
    oldPassword: string;
  }
}

/**
 * 修改当前用户密码。成功后后端会让此前签发的 token 全部失效，调用方需引导重新登录。
 */
export async function changePasswordApi(data: AuthApi.ChangePasswordParams) {
  return requestClient.post('/auth/change-password', data);
}

/**
 * 获取登录拼图验证码。拼图只能提交一次，失败后需重新获取。
 */
export async function getCaptchaApi() {
  return requestClient.get<AuthApi.CaptchaResult>('/auth/captcha');
}

/**
 * 提交拼图位置，通过时返回一次性登录凭证。
 * 走 baseRequestClient：失败（400/429）由拼图组件自己在图上提示，不弹全局错误消息。
 */
export async function verifyCaptchaApi(data: AuthApi.CaptchaVerifyParams) {
  return baseRequestClient.post<AuthApi.CaptchaVerifyRawResult>(
    '/auth/captcha/verify',
    data,
  );
}

/**
 * 登录
 */
export async function loginApi(data: AuthApi.LoginParams) {
  return requestClient.post<AuthApi.LoginResult>('/auth/login', data);
}

/**
 * 刷新accessToken
 */
export async function refreshTokenApi() {
  return baseRequestClient.post<AuthApi.RefreshTokenResult>('/auth/refresh', {
    withCredentials: true,
  });
}

/**
 * 退出登录
 */
export async function logoutApi() {
  return baseRequestClient.post('/auth/logout', {
    withCredentials: true,
  });
}

/**
 * 获取用户权限码
 */
export async function getAccessCodesApi() {
  return requestClient.get<string[]>('/auth/codes');
}
