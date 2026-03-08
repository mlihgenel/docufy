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

export type HomeFeature = {
  category: string;
  icon: HomeIcon;
  highlights: string[];
  transforms: string[];
  docHref: string;
};

export const capabilityFeatures: HomeFeature[] = [
  {
    category: "Yerel ve Guvenli Isleme",
    icon: "metadata",
    highlights: ["Dosyalarin cihaz disina cikmadan donusturulmesini saglar."],
    transforms: ["local-first isleme", "metadata preserve/strip kontrolu", "harici servise yukleme yok"],
    docHref: "/docs/nasil-calisir/"
  },
  {
    category: "Tek Aracta Cok Mod",
    icon: "automation",
    highlights: ["Tek dosya islemlerinden operasyonel akislara kadar ayni modelde calisir."],
    transforms: ["tek dosya donusum", "batch operasyon", "watch izleme", "pipeline akislari"],
    docHref: "/docs/rehberler/batch-ve-watch/"
  },
  {
    category: "Medya Araclari ve Duzenleme",
    icon: "video",
    highlights: ["Donusumun yaninda video/ses odakli yardimci araclar sunar."],
    transforms: ["trim (clip/remove)", "extract-audio ve snapshot", "video merge", "audio normalize"],
    docHref: "/docs/rehberler/video-ve-ses/"
  },
  {
    category: "Operasyonel Gorunurluk",
    icon: "normalize",
    highlights: ["Uzun surecli islerde takip, raporlama ve geri donus mekanizmasi saglar."],
    transforms: ["json rapor ciktilari", "resume-from-report", "on-conflict ve retry", "cli + tui deneyimi"],
    docHref: "/docs/hizli-baslangic/"
  }
];

export const conversionScopeFeatures: HomeFeature[] = [
  {
    category: "Belge ve Metin Kapsami",
    icon: "document",
    highlights: ["Ofis ve metin odakli donusum ihtiyaclarini kapsar."],
    transforms: ["md -> pdf", "docx -> txt", "pdf -> txt/html"],
    docHref: "/docs/rehberler/tek-dosya-donusturme/"
  },
  {
    category: "Gorsel Kapsami",
    icon: "image",
    highlights: ["Klasik raster formatlarin yaninda HEIC/HEIF/SVG kaynaklarini destekler."],
    transforms: ["jpg/png/webp arasi", "heic/heif -> png/jpg/webp/bmp/gif/tiff/ico", "svg -> raster/pdf"],
    docHref: "/docs/rehberler/tek-dosya-donusturme/"
  },
  {
    category: "Ses Kapsami",
    icon: "audio",
    highlights: ["Yaygin ses formatlari arasinda donusum ve normalize akislari sunar."],
    transforms: ["mp3 <-> wav", "m4a -> mp3/wav", "normalize cikti"],
    docHref: "/docs/rehberler/video-ve-ses/"
  },
  {
    category: "Video Kapsami",
    icon: "video",
    highlights: ["Video format donusumlerini ve gif cikti senaryolarini kapsar."],
    transforms: ["mp4/mov/webm -> mp4/gif", "video -> extract-audio"],
    docHref: "/docs/rehberler/tek-dosya-donusturme/"
  }
];

export const videoToolFeatures: HomeFeature[] = [
  {
    category: "Video Trim (Clip/Remove)",
    icon: "trim",
    highlights: ["Belirli araligi keser ya da kaldirir; plan odakli duzenleme sunar."],
    transforms: ["clip", "remove", "dry-run/preview"],
    docHref: "/docs/rehberler/video-ve-ses/"
  },
  {
    category: "Extract Audio",
    icon: "extract",
    highlights: ["Videodan ses kanalini ayri cikti olarak alir."],
    transforms: ["mp4/mov/webm -> mp3/wav"],
    docHref: "/docs/rehberler/video-ve-ses/"
  },
  {
    category: "Snapshot",
    icon: "snapshot",
    highlights: ["Videonun belirli anindan tek kare olusturur."],
    transforms: ["at second", "at %position", "jpg/png ciktisi"],
    docHref: "/docs/rehberler/video-ve-ses/"
  },
  {
    category: "Merge",
    icon: "merge",
    highlights: ["Birden fazla videoyu tek cikti halinde birlestirir."],
    transforms: ["multi-input", "reencode opsiyonu"],
    docHref: "/docs/rehberler/video-ve-ses/"
  }
];

export const extraMediaFeatures: HomeFeature[] = [
  {
    category: "Boyutlandirma ve Preset",
    icon: "resize",
    highlights: ["Manuel olcu ve hazir presetlerle medya boyutunu hedefe uyarlar."],
    transforms: ["story/square/fullhd", "pad/fit/fill/stretch"],
    docHref: "/docs/rehberler/tek-dosya-donusturme/"
  },
  {
    category: "Ses Normalizasyonu",
    icon: "normalize",
    highlights: ["LUFS hedefiyle ses seviyesini dengeler."],
    transforms: ["audio normalize", "target lufs"],
    docHref: "/docs/rehberler/video-ve-ses/"
  },
  {
    category: "Metadata Kontrolu",
    icon: "metadata",
    highlights: ["Ihtiyaca gore metadata koruma veya temizleme secenekleri sunar."],
    transforms: ["preserve", "strip"],
    docHref: "/docs/nasil-calisir/"
  },
  {
    category: "Batch / Watch / Pipeline",
    icon: "automation",
    highlights: ["Toplu ve surekli operasyonlar icin raporlama ve resume destekler."],
    transforms: ["batch", "watch", "pipeline", "resume/report"],
    docHref: "/docs/rehberler/batch-ve-watch/"
  }
];
