export interface AppError {
  message: string;
  code?: string;
  statusCode?: number;
  originalError?: unknown;
}

export class InsightError extends Error {
  public readonly code?: string;
  public readonly statusCode: number;
  public readonly originalError?: unknown;

  constructor(message: string, code?: string, statusCode = 500, originalError?: unknown) {
    super(message);
    this.name = 'InsightError';
    this.code = code;
    this.statusCode = statusCode;
    this.originalError = originalError;

    // TypeScriptのエラースタックトレースを保持
    if (Error.captureStackTrace) {
      Error.captureStackTrace(this, InsightError);
    }
  }
}

export function handleError(error: unknown, context: string): InsightError {
  if (error instanceof InsightError) {
    logError(error, context);
    return error;
  }

  if (error instanceof Error) {
    const appError = new InsightError(
      `${context}: ${error.message}`,
      'INTERNAL_ERROR',
      500,
      error
    );
    logError(appError, context);
    return appError;
  }

  const appError = new InsightError(
    `${context}: Unknown error occurred`,
    'UNKNOWN_ERROR',
    500,
    error
  );
  logError(appError, context);
  return appError;
}

export function logError(error: InsightError, context: string): void {
  const timestamp = new Date().toISOString();
  console.error(`[${timestamp}] [${context}] Error:`, {
    message: error.message,
    code: error.code,
    statusCode: error.statusCode,
    stack: error.stack,
    originalError: error.originalError
  });
}

export function createUserFriendlyMessage(error: InsightError): string {
  switch (error.code) {
    case 'DATABASE_ERROR':
      return 'データベースエラーが発生しました。しばらく時間をおいて再試行してください。';
    case 'AI_SERVICE_ERROR':
      return 'AI処理中にエラーが発生しました。しばらく時間をおいて再試行してください。';
    case 'VALIDATION_ERROR':
      return '入力データに問題があります。入力内容を確認してください。';
    case 'NOT_FOUND':
      return '指定されたリソースが見つかりませんでした。';
    default:
      return 'システムエラーが発生しました。しばらく時間をおいて再試行してください。';
  }
}