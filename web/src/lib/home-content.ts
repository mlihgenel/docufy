export type HomeIcon =
  | "document"
  | "image"
  | "audio"
  | "video"
  | "trim"
  | "extract"
  | "snapshot"
  | "merge"
  | "resize"
  | "normalize"
  | "metadata"
  | "automation";

export const homeIconPaths: Record<HomeIcon, string> = {
  document:
    "M8 3h7l4 4v14H8a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2Zm7 1.5V8h3.5M10 12h8M10 15h8M10 18h6",
  image:
    "M4 6a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6Zm4 9 2.6-3.2a1 1 0 0 1 1.56 0L14 14l1.4-1.7a1 1 0 0 1 1.55 0L19 15M9 9.5h.01",
  audio:
    "M5 16v-4a1 1 0 0 1 1-1h3l5-4v10l-5-4H6a1 1 0 0 1-1-1Zm11 3a5 5 0 0 0 0-14M18 16.5a2.5 2.5 0 0 0 0-7",
  video:
    "M5 7a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v1.2l3-2v11.6l-3-2V17a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V7Z",
  trim:
    "M5 7h5M5 17h5M11 6c2.8 0 5 2.2 5 5v6M11 18c2.8 0 5-2.2 5-5V7M5 12h6",
  extract:
    "M12 4v10m0 0-3-3m3 3 3-3M5 18h14M7 20h10",
  snapshot:
    "M4 8a2 2 0 0 1 2-2h2l1.2-2h5.6L16 6h2a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V8Zm8 9a3.5 3.5 0 1 0 0-7a3.5 3.5 0 0 0 0 7",
  merge:
    "M4 7h6l3 3m0 0 3-3h4M4 17h6l3-3m0 0 3 3h4",
  resize:
    "M4 9V4h5M20 9V4h-5M4 15v5h5M20 15v5h-5M9 12h6",
  normalize:
    "M5 15v-6m4 8V7m4 12V5m4 9v-4m-9 2h8",
  metadata:
    "M4 6a2 2 0 0 1 2-2h7l5 5v9a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6Zm8 0v4h4M8 15h8M8 12h4",
  automation:
    "M12 3v3M12 18v3M4.9 4.9l2.1 2.1M17 17l2.1 2.1M3 12h3M18 12h3M4.9 19.1 7 17M17 7l2.1-2.1M15.5 12a3.5 3.5 0 1 1-7 0a3.5 3.5 0 0 1 7 0Z"
};

export type CapabilityPillar = {
  title: string;
  icon: HomeIcon;
  summary: string;
  points: string[];
  docHref: string;
};

export type ConversionCoverage = {
  category: string;
  icon: HomeIcon;
  sources: string[];
  targets: string[];
  notes: string[];
  docHref: string;
};

export type ToolItem = {
  title: string;
  icon: HomeIcon;
  summary: string;
  points: string[];
  docHref: string;
};

export const capabilityPillars: CapabilityPillar[] = [
  {
    title: "Yerel ve Güvenli İşleme",
    icon: "metadata",
    summary: "Tüm dönüşümler cihazda çalışır; dosyalar üçüncü taraf servislere gönderilmez.",
    points: [
      "Yerel çalışma modeli (local-first)",
      "Metadata koruma veya temizleme seçenekleri",
      "Harici API'ye yükleme zorunluluğu yok"
    ],
    docHref: "/docs/nasil-calisir/"
  },
  {
    title: "Tek Araçta Çok Mod",
    icon: "automation",
    summary: "Tek dosya işlemlerinden operasyonel akışlara kadar aynı komut ailesiyle ilerlersin.",
    points: ["Tek dosya dönüşümü", "Batch operasyon", "Watch izleme", "Pipeline akışları"],
    docHref: "/docs/rehberler/batch-ve-watch/"
  },
  {
    title: "Medya Araçları ve Düzenleme",
    icon: "video",
    summary: "Dönüşümün yanında video/ses odaklı düzenleme ve yardımcı araçlar sunar.",
    points: ["Video trim (clip/remove)", "Extract-audio", "Snapshot", "Video merge", "Ses normalize"],
    docHref: "/docs/rehberler/video-ve-ses/"
  },
  {
    title: "Operasyonel Görünürlük ve Dayanıklılık",
    icon: "normalize",
    summary: "Uzun süren işlemlerde takip, raporlama ve kaldığı yerden devam mekanizmaları sağlar.",
    points: [
      "JSON/TXT rapor çıktıları",
      "--resume-from-report ile devam",
      "Çakışma yönetimi (overwrite/skip/versioned)",
      "Retry ve worker kontrolü"
    ],
    docHref: "/docs/hizli-baslangic/"
  }
];

export const conversionCoverage: ConversionCoverage[] = [
  {
    category: "Belge ve Metin Kapsamı",
    icon: "document",
    sources: ["md", "html", "pdf", "docx", "txt", "odt", "rtf", "csv"],
    targets: ["md", "html", "pdf", "docx", "txt", "odt", "rtf", "csv", "xlsx"],
    notes: ["Ek yol: csv → xlsx"],
    docHref: "/docs/rehberler/tek-dosya-donusturme/"
  },
  {
    category: "Görsel Kapsamı",
    icon: "image",
    sources: ["png", "jpg/jpeg", "webp", "bmp", "gif", "tif/tiff", "ico", "svg", "heic", "heif"],
    targets: ["png", "jpg/jpeg", "webp", "bmp", "gif", "tif/tiff", "ico", "pdf"],
    notes: ["Ek yol: svg → pdf", "HEIC/HEIF kaynakları desteklenir"],
    docHref: "/docs/rehberler/tek-dosya-donusturme/"
  },
  {
    category: "Ses Kapsamı",
    icon: "audio",
    sources: ["mp3", "wav", "ogg", "flac", "aac", "m4a", "wma", "opus", "webm"],
    targets: ["mp3", "wav", "ogg", "flac", "aac", "m4a", "wma", "opus", "webm"],
    notes: ["FFmpeg tabanlı dönüşüm", "audio normalize desteklenir"],
    docHref: "/docs/rehberler/video-ve-ses/"
  },
  {
    category: "Video Kapsamı",
    icon: "video",
    sources: ["mp4", "mov", "mkv", "avi", "webm", "m4v", "wmv", "flv"],
    targets: ["mp4", "mov", "mkv", "avi", "webm", "m4v", "wmv", "flv", "gif"],
    notes: ["Video → GIF", "Videodan ses çıkarma (extract-audio)"],
    docHref: "/docs/rehberler/tek-dosya-donusturme/"
  }
];

export const videoTools: ToolItem[] = [
  {
    title: "Video Trim (Clip / Remove)",
    icon: "trim",
    summary: "Belirli bir aralığı klip olarak çıkarır veya aralığı silip kalan parçaları birleştirir.",
    points: ["Clip modu", "Remove modu", "Dry-run / Preview planı"],
    docHref: "/docs/rehberler/video-ve-ses/"
  },
  {
    title: "Extract Audio",
    icon: "extract",
    summary: "Videodaki ses kanalını ayrı bir ses dosyası olarak çıkarır.",
    points: ["mp4/mov/webm kaynakları", "mp3/wav hedefleri", "Gerektiğinde copy modu"],
    docHref: "/docs/rehberler/video-ve-ses/"
  },
  {
    title: "Snapshot",
    icon: "snapshot",
    summary: "Videonun belirli anından tek kare görüntü üretir.",
    points: ["Saniye bazlı seçim", "Yüzde bazlı seçim", "jpg/png çıktısı"],
    docHref: "/docs/rehberler/video-ve-ses/"
  },
  {
    title: "Merge",
    icon: "merge",
    summary: "Birden fazla video dosyasını tek bir çıktı halinde birleştirir.",
    points: ["Çoklu giriş", "Re-encode opsiyonu", "İsimlendirme kontrolü"],
    docHref: "/docs/rehberler/video-ve-ses/"
  }
];

export const extraMediaTools: ToolItem[] = [
  {
    title: "Boyutlandırma ve Preset",
    icon: "resize",
    summary: "Görsel ve videoları manuel ölçü veya preset ile hedef boyuta uyumlar.",
    points: ["story/square/fullhd presetleri", "pad/fit/fill/stretch", "px ve cm birimleri"],
    docHref: "/docs/rehberler/tek-dosya-donusturme/"
  },
  {
    title: "Ses Normalizasyonu",
    icon: "normalize",
    summary: "Ses seviyesini hedef LUFS değerine göre dengeler.",
    points: ["audio normalize", "target LUFS", "true peak / LRA kontrolleri"],
    docHref: "/docs/rehberler/video-ve-ses/"
  },
  {
    title: "Metadata Kontrolü",
    icon: "metadata",
    summary: "İş akışına göre metadata bilgisini koruyabilir veya temizleyebilirsin.",
    points: ["--preserve-metadata", "--strip-metadata", "profil bazlı metadata modu"],
    docHref: "/docs/nasil-calisir/"
  },
  {
    title: "Batch / Watch / Pipeline",
    icon: "automation",
    summary: "Toplu ve sürekli operasyonları aynı modelde otomatikleştirir.",
    points: ["Batch + recursive", "Watch + polling fallback", "Pipeline JSON", "Rapor ve resume"],
    docHref: "/docs/rehberler/batch-ve-watch/"
  }
];

export type HomeFeature = {
  category: string;
  icon: HomeIcon;
  highlights: string[];
  transforms: string[];
  docHref: string;
};

export const capabilityFeatures: HomeFeature[] = capabilityPillars.map((item) => ({
  category: item.title,
  icon: item.icon,
  highlights: [item.summary],
  transforms: item.points,
  docHref: item.docHref
}));

export const conversionScopeFeatures: HomeFeature[] = conversionCoverage.map((item) => ({
  category: item.category,
  icon: item.icon,
  highlights: [],
  transforms: [
    `Kaynak: ${item.sources.join(", ")}`,
    `Hedef: ${item.targets.join(", ")}`,
    ...item.notes
  ],
  docHref: item.docHref
}));

export const videoToolFeatures: HomeFeature[] = videoTools.map((item) => ({
  category: item.title,
  icon: item.icon,
  highlights: [item.summary],
  transforms: item.points,
  docHref: item.docHref
}));

export const extraMediaFeatures: HomeFeature[] = extraMediaTools.map((item) => ({
  category: item.title,
  icon: item.icon,
  highlights: [item.summary],
  transforms: item.points,
  docHref: item.docHref
}));
