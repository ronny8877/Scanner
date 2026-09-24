export interface Valuation {
  score: number;
  tier: 'Ultra Premium' | 'High Value' | 'Brandable' | 'Standard' | string;
  estimatedMinUsd: number;
  estimatedMaxUsd: number;
  estimatedDisplay: string;
  lengthScore: number;
  tldScore: number;
  phoneticScore: number;
  keywordScore: number;
  highlights: string[];
}

export interface ScanResultItem {
  domain: string;
  rootName: string;
  tld: string;
  available: boolean;
  status: 'Available' | 'Registered';
  registeredAt?: string;
  registrar?: string;
  nameservers?: string[];
  valuation: Valuation;
  latencyMs: number;
}

export interface ScanReport {
  seedKeywords: string[];
  totalChecked: number;
  availableCount: number;
  takenCount: number;
  highValueCount: number;
  durationMs: number;
  items: ScanResultItem[];
}

export interface DNSRecords {
  a?: string[];
  aaaa?: string[];
  mx?: string[];
  ns?: string[];
  txt?: string[];
  cname?: string;
}

export interface DomainInquiry {
  domain: string;
  available: boolean;
  statusSummary: string;
  registeredAt?: string;
  updatedAt?: string;
  expiresAt?: string;
  domainAge?: string;
  daysToExpiry?: number;
  registrar?: string;
  registryHandle?: string;
  statusFlags?: string[];
  nameservers?: string[];
  dns: DNSRecords;
  valuation: Valuation;
  checkedAt: string;
  checkLatencyMs: number;
}

export interface PageInfo {
  url: string;
  path: string;
  title: string;
  description?: string;
  h1?: string;
  statusCode: number;
  contentType: string;
  latencyMs: number;
  depth: number;
  internalLinks: number;
  externalLinks: number;
  childrenPaths?: string[];
}

export interface SiteNode {
  segment: string;
  fullPath: string;
  title?: string;
  statusCode?: number;
  latencyMs?: number;
  internalOut?: number;
  externalOut?: number;
  children?: SiteNode[];
}

export interface CrawlReport {
  rootUrl: string;
  host: string;
  pagesCrawled: number;
  totalLinks: number;
  externalCount: number;
  techHeaders?: string[];
  durationMs: number;
  pages: PageInfo[];
  tree: SiteNode;
}
