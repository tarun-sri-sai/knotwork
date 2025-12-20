export const normalizeSearchParams = (params: {
  [key: string]: string | string[] | undefined;
}) => {
  const out: { [key: string]: string | undefined } = {};
  for (const key in params) {
    const value = params[key];
    out[key] = Array.isArray(value) ? value[0] : value;
  }

  return out;
};
