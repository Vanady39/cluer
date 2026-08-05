import { type ThemeConfig } from 'antd';

export const theme: ThemeConfig = {
  components: {
    // 👇 Input
    Input: {
      paddingBlock: 10,
      paddingInline: 16,
      borderRadius: 12,
      colorBgContainer: '#f5f5f5',
      colorBorder: '#f5f5f5',
      activeBorderColor: '#2f7efd',
      activeShadow: '0 0 0 3px rgba(47, 126, 253, 0.1)',
      colorTextPlaceholder: '#8c8c8c',
    },

    // 👇 Select
    Select: {
      controlHeight: 48,
      borderRadius: 12,
      colorBgContainer: '#f5f5f5',
      colorBorder: '#f5f5f5',
      activeBorderColor: '#2f7efd',
      colorTextPlaceholder: '#8c8c8c',
      boxShadow: 'none',
    },

    // 👇 Button
    Button: {
      borderRadius: 12,
      controlHeight: 48,
      contentFontSize: 16,
      primaryColor: '#ffffff',
      defaultBg: 'transparent',
      defaultBorderColor: 'transparent',
      defaultHoverBg: 'transparent',
      defaultHoverBorderColor: 'transparent',
      defaultHoverColor: '#ff4d4f',
      textHoverBg: 'transparent',
    },

    // 👇 Form
    Form: {
      itemMarginBottom: 24,
      labelFontSize: 14,
      labelHeight: 24,
      labelColor: '#1a1a1a',
    },

    // 👇 Upload
    Upload: {
      borderRadius: 12,
      colorBorder: '#d9d9d9',
    },

    // 👇 InputNumber
    InputNumber: {
      controlHeight: 48,
      borderRadius: 12,
      colorBgContainer: '#f5f5f5',
      colorBorder: '#f5f5f5',
      activeBorderColor: '#2f7efd',
      activeShadow: '0 0 0 3px rgba(47, 126, 253, 0.1)',
    },
  },

  token: {
    colorPrimary: '#2f7efd',
    borderRadius: 12,
    fontFamily: '-apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif',
    colorTextPlaceholder: '#8c8c8c',
    colorBgContainer: '#f5f5f5',
  },
};