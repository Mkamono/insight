export interface DatabaseConfig {
  url?: string;
}

export interface ServiceConfig {
  database: DatabaseConfig;
}