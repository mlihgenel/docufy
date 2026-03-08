export type DocNavItem = {
  title: string;
  href: string;
};

export type DocNavSection = {
  title: string;
  items: DocNavItem[];
};

export const docsNav: DocNavSection[] = [
  {
    title: "Temel",
    items: [
      { title: "Kurulum", href: "/docs/kurulum/" },
      { title: "Hızlı Başlangıç", href: "/docs/hizli-baslangic/" },
      { title: "Nasıl Çalışır", href: "/docs/nasil-calisir/" }
    ]
  },
  {
    title: "Rehberler",
    items: [
      { title: "Tek Dosya Dönüştürme", href: "/docs/rehberler/tek-dosya-donusturme/" },
      { title: "Batch ve Watch", href: "/docs/rehberler/batch-ve-watch/" },
      { title: "Pipeline", href: "/docs/rehberler/pipeline/" },
      { title: "Video ve Ses", href: "/docs/rehberler/video-ve-ses/" }
    ]
  },
  {
    title: "Destek",
    items: [{ title: "Sorun Giderme", href: "/docs/sorun-giderme/" }]
  }
];

export const isNavItemActive = (currentPath: string, href: string): boolean => {
  const current = currentPath.endsWith("/") ? currentPath : `${currentPath}/`;
  const target = href.endsWith("/") ? href : `${href}/`;
  return current === target;
};
