export type InstallationModel = {
  readonly version: string;
  readonly dev: boolean;
  readonly handle_lookup: boolean;
};

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null;

export const isInstallationModel = (value: unknown): value is InstallationModel =>
  isRecord(value) &&
  typeof value.version === 'string' &&
  typeof value.dev === 'boolean' &&
  typeof value.handle_lookup === 'boolean';
