export interface Valuation {
  score: number;
  tier: string;
  regFeeUsd?: number;
  regFeeDisplay?: string;
  estimatedMinUsd: number;
  estimatedMaxUsd: number;
  estimatedDisplay: string;
  lengthScore: number;
  tldScore: number;
  phoneticScore: number;
  keywordScore: number;
  isDictionaryWord?: boolean;
  highlights: string[];
}

export interface DictionaryPack {
  id: string;
  name: string;
  description: string;
  words: string[];
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
  dictionaryUsed?: string;
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
  cname?: string;
  ptr?: string[];
  mx?: string[];
  ns?: string[];
  txt?: string[];
  dmarc?: string[];
  spf?: string;
}

export interface DomainInquiry {
  domain: string;
  available: boolean;
  statusSummary: string;
  liveSiteUrl?: string;
  waybackCalendarUrl?: string;
  registeredAt?: string;
  updatedAt?: string;
  expiresAt?: string;
  domainAge?: string;
  daysToExpiry?: number;
  registrar?: string;
  registrarIana?: string;
  registryHandle?: string;
  dnssec?: string;
  statusFlags?: string[];
  nameservers?: string[];
  dns: DNSRecords;
  valuation: Valuation;
  checkedAt: string;
  checkLatencyMs: number;
}

export interface WaybackSnapshot {
  year: string;
  date: string;
  timestamp: string;
  archiveUrl: string;
  statusCode: string;
}

export interface HistoryTimeline {
  date: string;
  source: string;
  event: string;
  archiveUrl?: string;
}

export interface HistoryReport {
  domain: string;
  currentlyRegistered?: boolean;
  previouslyRegistered: boolean;
  historyVerdict: string;
  summaryNote: string;
  liveSiteUrl?: string;
  waybackCalendarUrl?: string;
  rdapCreatedDate?: string;
  rdapRegistrar?: string;
  firstSeenAt?: string;
  lastSeenAt?: string;
  firstSeenYear?: number;
  lastSeenYear?: number;
  totalSpanYears: number;
  waybackSnapshots: number;
  activeYears?: string[];
  snapshots?: WaybackSnapshot[];
  certCount: number;
  pastIssuers?: string[];
  pastSubdomains?: string[];
  milestones?: HistoryTimeline[];
  checkLatencyMs: number;
}

export interface DetectedTracker {
  name: string;
  category: 'AD_NETWORK' | 'ANALYTICS' | 'PIXEL_TRACKER' | 'SDK_TELEMETRY' | string;
  provider: string;
  matchedRule: string;
  riskLevel: 'LOW' | 'MODERATE' | 'HIGH' | string;
  description: string;
}

export interface TrackerTelemetry {
  privacyGrade: string;
  verdict: string;
  summary: string;
  adNetworksCount: number;
  analyticsCount: number;
  pixelsCount: number;
  telemetryCount: number;
  totalDetected: number;
  detectedTrackers?: DetectedTracker[];
}

export interface DetectedTech {
  name: string;
  category: 'FRAMEWORK' | 'UI_DESIGN' | 'BUILD_MOTION' | 'PLATFORM_AUTH' | 'EDGE_HOSTING' | string;
  confidence: 'CERTAIN' | 'HIGH' | 'DETECTED' | string;
  matchedBy: string;
  description: string;
}

export interface TechStackTelemetry {
  primaryFramework: string;
  uiSystemSummary: string;
  edgePlatform: string;
  summary: string;
  totalDetected: number;
  frameworksCount: number;
  uiCount: number;
  platformCount: number;
  infraCount: number;
  technologies?: DetectedTech[];
}

export interface CrawlerPermission {
  botName: string;
  category: string;
  status: 'ALLOWED' | 'PARTIAL' | 'BLOCKED' | string;
  matchedRule: string;
}

export interface RobotsAgentGroup {
  userAgent: string;
  disallow: string[];
  allow: string[];
  crawlDelay?: string;
}

export interface SitemapEntry {
  loc: string;
  path: string;
  lastMod?: string;
  ageLabel?: string;
  changeFreq?: string;
  priority?: string;
  isChildMap?: boolean;
}

export interface RobotsSitemapReport {
  targetUrl: string;
  host: string;
  checkedAt: string;
  durationMs: number;
  robotsFound: boolean;
  robotsUrl: string;
  robotsStatus: number;
  robotsSizeBytes: number;
  totalDisallowCount: number;
  totalAllowCount: number;
  declaredSitemaps?: string[];
  agentGroups?: RobotsAgentGroup[];
  botMatrix: CrawlerPermission[];
  rawRobotsPreview?: string;
  sitemapFound: boolean;
  sitemapUrl?: string;
  sitemapStatus?: number;
  isSitemapIndex?: boolean;
  childSitemaps?: SitemapEntry[];
  totalUrlsCount: number;
  newestLastMod?: string;
  oldestLastMod?: string;
  updatedLast7Days: number;
  updatedLast30Days: number;
  updatedLastYear: number;
  entries?: SitemapEntry[];
}

export interface MetaAuditCheck {
  id: string;
  label: string;
  status: 'PASS' | 'WARN' | 'MISSING' | string;
  details: string;
}

export interface MetaSocialReport {
  targetUrl: string;
  finalUrl: string;
  host: string;
  statusCode: number;
  durationMs: number;
  checkedAt: string;
  title: string;
  description: string;
  canonicalUrl?: string;
  faviconUrl?: string;
  themeColor?: string;
  robotsMeta?: string;
  author?: string;
  generator?: string;
  language?: string;
  charset?: string;
  viewport?: string;
  ogTitle?: string;
  ogDescription?: string;
  ogImage?: string;
  ogUrl?: string;
  ogSiteName?: string;
  ogType?: string;
  ogLocale?: string;
  twitterCard?: string;
  twitterTitle?: string;
  twitterDescription?: string;
  twitterImage?: string;
  twitterSite?: string;
  twitterCreator?: string;
  resolvedTitle: string;
  resolvedDescription: string;
  resolvedImage?: string;
  resolvedSiteName: string;
  resolvedThemeColor: string;
  socialScore: number;
  socialGrade: string;
  auditChecks: MetaAuditCheck[];
  allMetaTags?: Record<string, string>;
  trackers: TrackerTelemetry;
  techStack?: TechStackTelemetry;
}

export interface PortProbe {
  port: number;
  service: string;
  protocol: string;
  open: boolean;
  latencyMs: number;
  category: string;
  riskNote?: string;
}

export interface TLSInfo {
  supported: boolean;
  version?: string;
  cipherSuite?: string;
  issuer?: string;
  subject?: string;
  validFrom?: string;
  validUntil?: string;
  daysRemaining: number;
  sans?: string[];
}

export interface SubdomainHit {
  subdomain: string;
  ips: string[];
  cname?: string;
  takeoverRisk?: string;
}

export interface SecurityCheck {
  control: string;
  passed: boolean;
  detail: string;
}

export interface ReconReport {
  domain: string;
  liveSiteUrl?: string;
  waybackCalendarUrl?: string;
  targetIp: string;
  reversePtr?: string;
  openPortsCount: number;
  portsScanned: number;
  ports: PortProbe[];
  tls: TLSInfo;
  subdomains: SubdomainHit[];
  httpHeaders?: Record<string, string>;
  hasRobotsTxt?: boolean;
  hasSecurityTxt?: boolean;
  hasSitemapXml?: boolean;
  securityGrade: string;
  securityScore: number;
  securityChecks: SecurityCheck[];
  trackers?: TrackerTelemetry;
  techStack?: TechStackTelemetry;
  durationMs: number;
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
  seedPath?: string;
  host: string;
  pagesCrawled: number;
  totalLinks: number;
  externalCount: number;
  techHeaders?: string[];
  durationMs: number;
  pages: PageInfo[];
  tree: SiteNode;
  trackers?: TrackerTelemetry;
  techStack?: TechStackTelemetry;
}

export interface SavedDomain {
  domain: string;
  available: boolean;
  status: string;
  valuation: Valuation;
  notes?: string;
  tags?: string[];
  previouslyRegistered?: boolean;
  firstSeenYear?: number;
  savedAt: string;
  lastCheckedAt: string;
}

export interface Job {
  id: string;
  type: 'scan' | 'inspect' | 'history' | 'recon' | 'crawl' | 'robots' | 'meta' | 'parallel_suite' | 'watchlist_recheck' | string;
  title: string;
  target: string;
  status: 'QUEUED' | 'RUNNING' | 'COMPLETED' | 'CANCELED' | 'CANCELLED' | 'FAILED';
  progress: number;
  phase: string;
  workers: number;
  resultSummary: string;
  createdAt: string;
  durationMs: number;
  result?: unknown;
}

export interface RankPoint {
  date: string;
  rank: number;
}

export interface TrafficSignal {
  source: string;
  value: string;
  weight: string;
  description: string;
}

export interface TrafficReport {
  domain: string;
  checkedAt: string;
  durationMs: number;
  isRegistered: boolean;
  isRanked: boolean;
  trancoRank: number;
  popularityTier: string;
  cloudflareBucket: string;
  estimatedMonthlyRange: string;
  estimatedDailyRange: string;
  confidenceLevel: string;
  trendDirection: 'RISING' | 'STABLE' | 'COOLING' | string;
  trendLabel: string;
  rankDelta30d: number;
  topLocations: string[];
  rankHistory?: RankPoint[];
  sitemapPagesCount: number;
  subdomainCount: number;
  waybackYearsCount: number;
  edgeNetwork: string;
  signals: TrafficSignal[];
  methodologyNote: string;
}

export interface ParallelSuiteResult {
  domain: string;
  generatedAt?: string;
  inquiry: DomainInquiry;
  history: HistoryReport;
  traffic: TrafficReport;
  recon: ReconReport;
  crawl: CrawlReport;
  robots: RobotsSitemapReport;
  meta: MetaSocialReport;
}

export interface SavedExecutiveReport {
  id: string;
  domain: string;
  generatedAt: string;
  suite: ParallelSuiteResult;
}

