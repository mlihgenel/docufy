# Docufy

<p align="center">
  <img src="docs/assets/docufy.gif" alt="Docufy Arayüzü">
</p> 



<p align="center">
  Belgeleri, görselleri, sesleri ve videoları tamamen yerel ortamda dönüştüren modern bir CLI/TUI aracı.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square" alt="License">
  <img src="https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-blue?style=flat-square" alt="Platform">
  <a href="https://goreportcard.com/report/github.com/mlihgenel/docufy"><img src="https://goreportcard.com/badge/github.com/mlihgenel/docufy?style=flat-square" alt="Go Report Card"></a>
</p>

## İçindekiler
- [Genel Bakış](#genel-bakış)
- [Özellikler](#özellikler)
- [Kurulum](#kurulum)
- [Hızlı Başlangıç](#hızlı-başlangıç)
- [Shell Completion](#shell-completion)
- [Komut Referansı](#komut-referansı)
- [Flag Referansı](#flag-referansı)
- [Pipeline Spec (JSON)](#pipeline-spec-json)
- [Desteklenen Formatlar](#desteklenen-formatlar)
- [Harici Bağımlılıklar](#harici-bağımlılıklar)
- [Yapılandırma](#yapılandırma)
- [Sorun Giderme](#sorun-giderme)
- [Release](#release)
- [Geliştirme](#geliştirme)
- [Proje Yapısı](#proje-yapısı)
- [Katkı](#katkı)
- [Lisans](#lisans)

## Genel Bakış
Docufy, dosya dönüştürme işlemlerini internet servislerine yükleme yapmadan yerel makinede gerçekleştiren bir komut satırı uygulamasıdır.

- Gizlilik odaklıdır: dosyalar cihazdan çıkmaz.
- İki kullanım modu sunar: CLI (otomasyon/script) ve interaktif TUI (menü tabanlı).

## Özellikler
- Belge, görsel, ses ve video dönüşümleri.
- Sabit görevlerde ses ve video manipülasyonu: videodan ses çıkarma (`extract-audio`), belirli anından kare yakalama (`snapshot`), videoları sıralı birleştirme (`merge`) ve ses dizeleme (`audio normalize`).
- WebP encode desteği: tüm görsel formatlarından WebP'ye dönüşüm (pure Go, lossless VP8L).
- HEIC/HEIF kaynak desteği: iPhone görsellerini PNG/JPG/WEBP/BMP/GIF/TIFF/ICO formatlarına dönüştürme.
- SVG kaynak desteği: SVG dosyalarını raster görsellere ve PDF çıktısına dönüştürme.
- Görsel optimizasyon: `--optimize` ile dosya boyutunu minimize etme, `--target-size 500kb` ile hedef boyuta yaklaşma.
- Dosya bilgisi komutu: `info` ile format, çözünürlük, codec, süre, bitrate bilgisi (JSON çıktı desteği).
- `mp4 -> gif` dahil video dönüşümü.
- Uzun FFmpeg işlemlerinde gerçek zamanlı progress bar ve ETA gösterimi (CLI, TUI ve batch).
- Video düzenleme (`video trim`): `clip` modunda aralık çıkarır, `remove` modunda aralığı silip kalan parçaları birleştirir.
- Video trim preview/plan: CLI’de `--dry-run/--preview`; TUI’de çalıştırmadan önce plan onayı ekranı.
- Video trim codec stratejisi: `--codec auto` (varsayılan) hedef formata göre uyumlu codec seçer.
- TUI video trim timeline adımı: başlangıç/bitiş aralığını klavye ile hızlı kaydırma ve remove modunda çoklu segment yönetimi (`a/n/p/d/m`).
- Görsel/video boyutlandırma: manuel (`px`/`cm`) veya hazır preset (`story`, `square`, `fullhd` vb.).
- Oranı koruyarak dikey/yatay uyarlama (`pad`, `fit`, `fill`, `stretch`); `pad` modunda siyah boşluk desteği.
- Interaktif ana menüde ayrı akışlar: `Dosya Dönüştür`, `Toplu Dönüştür`, `Klasör İzle`, `Video Düzenle (Klip/Sil)`, `Boyutlandır`, `Toplu Boyutlandır`, `Dosya Bilgisi`.
- Batch dönüşüm (dizin veya glob pattern).
- Paralel işleme (`--workers`) ile yüksek performans.
- Ön izleme modu (`--dry-run`) ile risksiz batch planlama.
- Çıktı dizinine yazarken klasör yapısını koruma (`batch --preserve-tree`).
- Çakışma politikası (`--on-conflict`: `overwrite`, `skip`, `versioned`).
- Otomatik retry (`--retry`, `--retry-delay`) ve raporlama (`--report`, `--report-file`).
- Hazır + kullanıcı tanımlı profil sistemi (`--profile`: built-in profiller ve `~/.docufy/profiles/*.toml`).
- Metadata kontrolü (`--preserve-metadata`, `--strip-metadata`).
- Klasör izleme ile otomatik dönüşüm (`watch` komutu, event-driven + polling fallback).
- Makine-okunur CLI çıktısı (`--output-format json`).
- Proje bazlı ayarlar: `.docufy.toml` (flag > env > project config > default).
- Harici bağımlılık kontrolü (FFmpeg, LibreOffice, Pandoc).
- Format alias desteği (`jpeg -> jpg`, `tiff -> tif`, `markdown -> md`).

## Kurulum

### 1. Go ile kurulum (önerilen)
```bash
go install github.com/mlihgenel/docufy/v2/cmd/docufy@latest
```

Kurulum sonrası herhangi bir dizinden çalıştırabilmek için binary yolunun `PATH` içinde olması gerekir.
`go env GOBIN` doluysa o dizini, boşsa `$(go env GOPATH)/bin` dizinini `PATH` içine ekleyin.

### 2. PATH ayarı (herhangi bir dizinden çalıştırmak için)

#### macOS / Linux (zsh veya bash)
```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
```

`bash` kullanıyorsanız `~/.bashrc` veya `~/.bash_profile` dosyasına ekleyin.

#### Windows (PowerShell)
```powershell
$gopath = go env GOPATH
setx PATH "$env:PATH;$gopath\bin"
```

Ardından yeni bir terminal açın.

### 3. Kaynaktan derleme
```bash
git clone https://github.com/mlihgenel/docufy.git
cd docufy
go build -o docufy ./cmd/docufy
./docufy --help
```

Not: Sürüm bilgisi artık build metadata'dan otomatik okunur; `cmd/docufy/main.go` içinde elle sürüm güncellemek gerekmez.
Release için isterseniz sürümü build anında net verebilirsiniz:
```bash
go build -ldflags "-X main.version=$(git describe --tags --always --dirty | sed 's/^v//')" -o docufy ./cmd/docufy
```

Windows için:
```powershell
go build -o docufy.exe ./cmd/docufy
.\docufy.exe --help
```

## Hızlı Başlangıç

### Yardım menüsü
```bash
docufy --help
docufy help convert
docufy help batch
docufy help watch
docufy help pipeline
docufy help video
docufy help formats
docufy help resize-presets
docufy help completion
docufy help profiles
```

### İnteraktif mod (TUI)
```bash
docufy
```

Interaktif ana menü (bölüm bazlı):
- `Dönüştürme`: tek dosya, toplu ve watch akışları
- `Video Araçları`: klip çıkarma ve aralık silme + birleştirme (`başlangıç + süre` ya da `başlangıç + bitiş`)
- `Boyutlandırma`: tek dosya ve toplu boyutlandırma
- `Bilgi ve Ayarlar`: desteklenen formatlar, sistem kontrolü, ayarlar

TUI açmadan doğrudan CLI ile çalışmak için:
```bash
docufy --help
docufy help <komut>
```

### Shell completion
```bash
# Zsh
docufy completion zsh > "${fpath[1]}/_docufy"

# Bash
docufy completion bash > /etc/bash_completion.d/docufy
```

Detaylar ve diğer shell'ler için:
```bash
docufy completion --help
```

### Format sorgulama
```bash
docufy formats
docufy formats --from mp4
docufy formats --to gif
docufy formats --output-format json
```

### Tek dosya dönüşümü
```bash
# Belge
docufy convert belge.md --to pdf

# Görsel
docufy convert fotograf.jpeg --to png

# HEIC/HEIF görseli PNG'ye dönüştür
docufy convert IMG_1234.HEIC --to png

# SVG dosyasını PDF'e dönüştür
docufy convert logo.svg --to pdf

# Görseli WebP'ye dönüştür
docufy convert fotograf.png --to webp

# Görsel optimizasyonu (dosya boyutunu küçült)
docufy convert fotograf.jpg --to jpg --optimize
docufy convert fotograf.jpg --to jpg --target-size 500kb

# Ses
docufy convert ses.mp3 --to wav

# Video -> GIF
docufy convert klip.mp4 --to gif --quality 80

# Yatay videoyu dikeye çevir (siyah boşluklarla oran koru)
docufy convert klip.mp4 --to mp4 --preset story --resize-mode pad

# Görseli manuel ölçüyle boyutlandır (cm)
docufy convert fotograf.webp --to png --width 12 --height 18 --unit cm --dpi 300

# Profil kullanımı (story çıktı için)
docufy convert klip.mp4 --to mp4 --profile social-story

# Kullanıcı profili oluştur ve kullan
docufy profiles create story-fast --quality 83 --preset story --resize-mode fit --metadata-mode strip
docufy convert klip.mp4 --to mp4 --profile story-fast

# Metadata temizleme
docufy convert kamera.mov --to mp4 --strip-metadata

# Dosya bilgisi görme
docufy info fotograf.jpg
docufy info video.mp4 --output-format json
```

### Toplu (batch) dönüşüm
```bash
# Dizindeki tüm .md dosyalarını PDF yap
docufy batch ./docs --from md --to pdf

# Alt dizinlerle birlikte
docufy batch ./videolar --from mp4 --to gif --recursive

# Ön izleme (dönüştürmeden planı gösterir)
docufy batch ./resimler --from jpg --to png --dry-run

# Glob kullanımı
docufy batch "*.png" --from png --to jpg --quality 85

# Toplu olarak story ölçüsüne getir
docufy batch ./videolar --from mp4 --to mp4 --preset story --resize-mode pad

# Recursive batch'te çıktı klasörü altında kaynak dizin yapısını koru
docufy batch ./assets --from png --to jpg --recursive --output ./export --preserve-tree

# Çakışma ve retry ile JSON rapor üret
docufy batch ./resimler --from jpg --to png --on-conflict versioned --retry 2 --retry-delay 1s --report json --report-file ./reports/batch.json

# Önceki JSON raporundan başarılı işleri atlayarak devam et
docufy batch ./resimler --from jpg --to png --resume-from-report ./reports/batch.json

# Profil + metadata modu ile batch
docufy batch ./videolar --from mp4 --to mp4 --profile social-story --strip-metadata
```

### Watch modu (otomatik dönüşüm)
```bash
# incoming klasörünü izle, yeni webp dosyalarını jpg yap
docufy watch ./incoming --from webp --to jpg

# Alt dizinlerle birlikte izle
docufy watch ./videolar --from mp4 --to gif --recursive --quality 80

# Profil ile izle
docufy watch ./incoming --from mov --to mp4 --profile archive-lossless
```

### Pipeline modu (çok adımlı akış)
```bash
# Pipeline spec dosyasını çalıştır
docufy pipeline run ./pipeline.json

# Profil ve metadata ile çalıştır, JSON rapor al
docufy pipeline run ./pipeline.json --profile social-story --strip-metadata --report json --report-file ./reports/pipeline.json

# Önceki JSON rapora göre başarılı step'leri atlayıp kaldığı yerden devam et
docufy pipeline run ./pipeline.json --resume-from-report ./reports/pipeline.json
```

Örnek spec dosyası: `examples/pipeline.example.json`

### Pipeline Spec (JSON)

| Alan | Zorunlu | Açıklama |
|---|---|---|
| `input` | Evet | Pipeline'ın başlangıç dosyası |
| `output` | Hayır | Son adımın nihai çıktı yolu |
| `steps[]` | Evet | Sıralı işlem adımları |
| `steps[].type` | Evet | `convert`, `extract-audio` veya `audio-normalize` |
| `steps[].to` | `convert` ve `extract-audio` için evet | Hedef format (`mp3`, `wav`, `pdf` vb.) |
| `steps[].quality` | Hayır | Adım bazlı kalite (1-100) |
| `steps[].output` | Hayır | O adım için özel çıktı yolu |
| `steps[].metadata_mode` | Hayır | `auto`, `preserve`, `strip` |
| `steps[].target_lufs` | `audio-normalize` için hayır | Hedef LUFS |
| `steps[].target_tp` | `audio-normalize` için hayır | Hedef true peak |
| `steps[].target_lra` | `audio-normalize` için hayır | Hedef loudness range |

### Video ve Ses Araçları
```bash
# Videodan ses kanalını MP3 olarak çıkar
docufy video extract-audio klip.mp4

# Videodan ses kanalını re-encode etmeden (copy) orijinal formatıyla çıkar
docufy video extract-audio orijinal.mov --copy

# Videonun 30. saniyesinden tek kare (snapshot) al
docufy video snapshot klip.mp4 --at 30 --to jpg

# Videonun tam ortasından (%50) yüksek kalite snapshot al
docufy video snapshot klip.mp4 --at %50 --to png

# Aynı codec'e sahip parçaları hızlıca birleştir (concat demuxer)
docufy video merge part1.mp4 part2.mp4 --name full_video

# Farklı codec'lere sahip videoları re-encode ederek birleştir
docufy video merge iphone.mov web.webm --to mp4 --reencode --quality 80

# Ses dosyasının ses seviyesini EBU R128 (LUFS) standardına göre normalize et
docufy audio normalize podcast.mp3 --target-lufs -16

# 5. saniyeden başlayıp 10 saniyelik klip çıkar
docufy video trim input.mp4 --start 00:00:05 --duration 10

# 23-25 saniye aralığını videodan sil ve kalan parçaları birleştir
docufy video trim input.mp4 --mode remove --start 00:00:23 --duration 2

# Birden fazla aralığı tek seferde sil (sadece remove modunda)
docufy video trim input.mp4 --mode remove --ranges "00:00:05-00:00:08,00:00:20-00:00:25"

# Preview/plan: işlemden önce tam etkiyi gör (dosya yazmaz)
docufy video trim input.mp4 --mode remove --ranges "5-8,20-25" --dry-run
```

## Komut Referansı

| Komut | Ne yapar | Örnek |
|---|---|---|
| `docufy` | İnteraktif TUI modunu başlatır | `docufy` |
| `docufy convert <dosya>` | Tek dosya dönüşümü | `docufy convert input.mp4 --to gif` |
| `docufy batch <dizin/glob>` | Toplu dönüşüm | `docufy batch ./src --from md --to html` |
| `docufy watch <dizin>` | Klasörü izleyip otomatik dönüşüm yapar | `docufy watch ./incoming --from webp --to jpg` |
| `docufy pipeline run <dosya>` | JSON pipeline akışını çalıştırır | `docufy pipeline run ./pipeline.json` |
| `docufy video trim <dosya>` | `clip`: aralık çıkarır, `remove`: aralığı siler + birleştirir | `docufy video trim input.mp4 --mode remove --start 00:00:23 --duration 2` |
| `docufy video extract-audio <dosya>` | Videodan ses kanalını çıkarır | `docufy video extract-audio input.mp4 --to wav` |
| `docufy video snapshot <dosya>` | Videodan tek kare seçer | `docufy video snapshot input.mp4 --at %50` |
| `docufy video merge <dosyalar...>` | Birden fazla videoyu birleştirir | `docufy video merge part1.mp4 part2.mp4` |
| `docufy audio normalize <dosya>` | Ses seviyesini EBU R128'e göre dengeler | `docufy audio normalize ses.mp3 --target-lufs -14` |
| `docufy resize-presets` | Hazır boyut presetlerini listeler | `docufy resize-presets` |
| `docufy info <dosya>` | Dosya bilgisi gösterir (format, boyut, çözünürlük, codec) | `docufy info foto.jpg` |
| `docufy formats` | Desteklenen dönüşümleri listeler | `docufy formats --from pdf` |
| `docufy completion <shell>` | Shell completion üretir | `docufy completion zsh` |
| `docufy profiles list` | Built-in ve kullanıcı profillerini listeler | `docufy profiles list` |
| `docufy profiles create [ad]` | Yeni kullanıcı profili oluşturur | `docufy profiles create story-fast --quality 83` |
| `docufy help [komut]` | Komut yardımı gösterir | `docufy help batch` |

## Flag Referansı

### Global flag'ler

| Flag | Kısa | Açıklama |
|---|---|---|
| `--output` | `-o` | Çıktı dizini (varsayılan: kaynak dosya dizini) |
| `--verbose` | `-v` | Detaylı çıktı |
| `--workers` | `-w` | Batch modunda paralel worker sayısı |
| `--output-format` | - | CLI çıktı formatı: `text` veya `json` |
| `--version` | - | Sürüm bilgisini gösterir |

### `convert` flag'leri

| Flag | Kısa | Açıklama |
|---|---|---|
| `--to` | `-t` | Hedef format (zorunlu) |
| `--profile` | - | Profil adı: built-in veya `~/.docufy/profiles/` altındaki kullanıcı profili |
| `--quality` | `-q` | Kalite seviyesi (1-100) |
| `--name` | `-n` | Çıktı dosya adı (uzantısız) |
| `--on-conflict` | - | Çakışma politikası: `overwrite`, `skip`, `versioned` |
| `--preserve-metadata` | - | Metadata bilgisini korumayı dener |
| `--strip-metadata` | - | Metadata bilgisini temizler |
| `--preset` | - | Hazır boyut (ör: `story`, `square`, `fullhd`, `1080x1920`) |
| `--width` | - | Manuel genişlik değeri |
| `--height` | - | Manuel yükseklik değeri |
| `--unit` | - | Manuel birim (`px` veya `cm`) |
| `--dpi` | - | `cm` kullanıldığında DPI değeri |
| `--resize-mode` | - | Boyutlandırma modu: `pad`, `fit`, `fill`, `stretch` |
| `--optimize` | - | Dosya boyutunu minimize et (görsel dönüşümlerinde) |
| `--target-size` | - | Hedef dosya boyutu (ör: `500kb`, `2mb`) |

### `batch` flag'leri

| Flag | Kısa | Açıklama |
|---|---|---|
| `--from` | `-f` | Kaynak format (zorunlu) |
| `--to` | `-t` | Hedef format (zorunlu) |
| `--profile` | - | Profil adı: built-in veya `~/.docufy/profiles/` altındaki kullanıcı profili |
| `--recursive` | `-r` | Alt dizinleri de tara |
| `--preserve-tree` | - | Dizin modunda `--output` altına kaynak klasör yapısını korur |
| `--dry-run` | - | Dönüştürmeden önce planı göster |
| `--quality` | `-q` | Kalite seviyesi (1-100) |
| `--on-conflict` | - | Çakışma politikası: `overwrite`, `skip`, `versioned` |
| `--preserve-metadata` | - | Metadata bilgisini korumayı dener |
| `--strip-metadata` | - | Metadata bilgisini temizler |
| `--retry` | - | Başarısız işler için otomatik tekrar sayısı |
| `--retry-delay` | - | Retry denemeleri arası bekleme (`500ms`, `2s` vb.) |
| `--report` | - | Rapor formatı: `off`, `txt`, `json` |
| `--report-file` | - | Raporu belirtilen dosyaya yazar |
| `--resume-from-report` | - | Önceki JSON rapordaki `success` girdileri atlayarak devam eder |
| `--preset` | - | Hazır boyut (ör: `story`, `square`, `fullhd`, `1080x1920`) |
| `--width` | - | Manuel genişlik değeri |
| `--height` | - | Manuel yükseklik değeri |
| `--unit` | - | Manuel birim (`px` veya `cm`) |
| `--dpi` | - | `cm` kullanıldığında DPI değeri |
| `--resize-mode` | - | Boyutlandırma modu: `pad`, `fit`, `fill`, `stretch` |

### `watch` flag'leri

| Flag | Kısa | Açıklama |
|---|---|---|
| `--from` | `-f` | Kaynak format (zorunlu) |
| `--to` | `-t` | Hedef format (zorunlu) |
| `--profile` | - | Profil adı: built-in veya `~/.docufy/profiles/` altındaki kullanıcı profili |
| `--recursive` | `-r` | Alt dizinleri de izle |
| `--quality` | `-q` | Kalite seviyesi (1-100) |
| `--on-conflict` | - | Çakışma politikası: `overwrite`, `skip`, `versioned` |
| `--preserve-metadata` | - | Metadata bilgisini korumayı dener |
| `--strip-metadata` | - | Metadata bilgisini temizler |
| `--retry` | - | Başarısız işler için otomatik tekrar sayısı |
| `--retry-delay` | - | Retry denemeleri arası bekleme (`500ms`, `2s` vb.) |
| `--interval` | - | Periyodik tarama aralığı (event modunda fallback/sağlık kontrolü) |
| `--settle` | - | Dosyanın stabil sayılması için bekleme süresi |

### `pipeline run` flag'leri

| Flag | Kısa | Açıklama |
|---|---|---|
| `--profile` | - | Profil adı: built-in veya `~/.docufy/profiles/` altındaki kullanıcı profili |
| `--quality` | `-q` | Varsayılan kalite seviyesi (1-100) |
| `--on-conflict` | - | Çakışma politikası: `overwrite`, `skip`, `versioned` |
| `--preserve-metadata` | - | Metadata bilgisini korumayı dener |
| `--strip-metadata` | - | Metadata bilgisini temizler |
| `--report` | - | Rapor formatı: `off`, `txt`, `json` |
| `--report-file` | - | Raporu belirtilen dosyaya yazar |
| `--resume-from-report` | - | Önceki JSON pipeline raporuna göre başarılı step'leri atlayıp devam eder |
| `--keep-temps` | - | Ara geçici dosyaları silmez |

### `video trim` flag'leri

| Flag | Kısa | Açıklama |
|---|---|---|
| `--mode` | - | İşlem modu: `clip` veya `remove` |
| `--start` | - | İşlem başlangıç zamanı (örn: `00:00:05`) |
| `--end` | - | Bitiş zamanı (`--duration` ile birlikte kullanılamaz) |
| `--duration` | - | İşlem süresi (örn: `10`, `00:00:10`) |
| `--ranges` | - | Sadece `remove` modunda çoklu aralık listesi (örn: `00:00:05-00:00:08,00:00:20-00:00:25`) |
| `--dry-run` | - | İşlem yapmadan plan/etki ön izlemesi gösterir |
| `--preview` | - | `--dry-run` ile aynı davranış |
| `--codec` | - | `auto` (önerilen), `copy`, `reencode` |
| `--to` | - | Hedef format (`mp4`, `mov` vb.) |
| `--output-file` | - | Tam çıktı dosya yolu |
| `--name` | `-n` | Çıktı dosya adı (uzantısız) |
| `--profile` | - | Profil adı: built-in veya `~/.docufy/profiles/` altındaki kullanıcı profili |
| `--quality` | `-q` | Reencode modunda kalite seviyesi |
| `--on-conflict` | - | Çakışma politikası: `overwrite`, `skip`, `versioned` |
| `--preserve-metadata` | - | Metadata bilgisini korumayı dener |
| `--strip-metadata` | - | Metadata bilgisini temizler |

### `formats` flag'leri

| Flag | Açıklama |
|---|---|
| `--from` | Belirli bir kaynaktan gidilebilen hedefleri listeler |
| `--to` | Belirli bir hedefe gelebilen kaynakları listeler |

### Boyutlandırma modları
- `pad`: Oranı korur, hedef boyutu doldurmak için siyah boşluk ekler (yatay -> dikey için önerilen).
- `fit`: Oranı korur, hedef kutuya sığdırır; çıktı bir kenarda daha küçük kalabilir.
- `fill`: Oranı korur, hedef kutuyu doldurur; taşan kısmı ortadan kırpar.
- `stretch`: Oranı korumaz, hedef ölçüye zorla esnetir.

### Profiller
Profil, dönüştürme komutlarında tekrar eden ayarları tek isimle uygulayan bir "ayar preset" mekanizmasıdır.
Format secimi yapmaz; sadece `quality`, `retry`, `metadata`, `resize`, `on-conflict` gibi bayraklara varsayılan deger verir.

Built-in profiller kapsam alanina gore gelir:
- `Gorsel+Video`: `social-story`, `social-feed`
- `Video`: `social-reel-fast`, `video-web-balanced`, `video-compress-fast`
- `Gorsel`: `image-web-balanced`, `image-print-a4`, `image-archive`
- `Ses`: `podcast-clean`, `voice-note-fast`
- `Belge`: `doc-share`, `doc-compact`
- `Genel`: `archive-lossless`

Not:
- Profil kapsam disi bir kaynakta kullanilirsa CLI net hata verir.
- Komutta verdiginiz bir flag, profildeki ayni degerin ustune yazar.

Kullanıcı profilleri:
- Dizin: `~/.docufy/profiles/`
- Format: TOML
- Komutlar: `docufy profiles list` ve `docufy profiles create`
- Aynı isimli kullanıcı profili, built-in profilin alanlarını override eder.
- `scope` alani: `all`, `document`, `image`, `video`, `audio` veya birden fazla (`image,video`).

Örnek kullanıcı profili:
```toml
description = "Story ciktilari icin hizli varsayilanlar"
scope = "image,video"
quality = 83
on_conflict = "versioned"
resize_preset = "story"
resize_mode = "fit"
metadata_mode = "strip"
retry = 1
retry_delay = "500ms"
```

Örnek dosya yolu:
```text
~/.docufy/profiles/story-fast.toml
```

Kullanım:
```bash
docufy profiles list
docufy profiles create story-fast --scope image,video --quality 83 --preset story --resize-mode fit --metadata-mode strip
docufy convert klip.mp4 --to mp4 --profile story-fast
docufy batch ./videolar --from mov --to mp4 --profile story-fast
docufy watch ./incoming --from mov --to mp4 --profile story-fast
```

## Desteklenen Formatlar

En güncel ve tam matris için:
```bash
docufy formats
```

### Belgeler
- Kaynak/hedef: `md`, `html`, `pdf`, `docx`, `txt`, `odt`, `rtf`, `csv`
- Ek: `csv -> xlsx`

### Görseller
- Kaynak: `png`, `jpg/jpeg`, `webp`, `bmp`, `gif`, `tif/tiff`, `ico`, `svg`, `heic`, `heif`
- Hedef: `png`, `jpg/jpeg`, `webp`, `bmp`, `gif`, `tif/tiff`, `ico`
- Ek: `svg -> pdf`

### Ses (FFmpeg)
- `mp3`, `wav`, `ogg`, `flac`, `aac`, `m4a`, `wma`, `opus`, `webm`

### Videolar (FFmpeg)
- Kaynak: `mp4`, `mov`, `mkv`, `avi`, `webm`, `m4v`, `wmv`, `flv`
- Hedef: yukarıdakiler + `gif`

## Harici Bağımlılıklar

| Araç | Ne zaman gerekir | Not |
|---|---|---|
| FFmpeg | Ses ve video dönüşümleri | `mp4 -> gif` dahil |
| LibreOffice | Bazı belge dönüşümleri (`odt/rtf/xlsx`) | Bazı dönüşümler için fallback kullanılır |
| Pandoc | Bazı Markdown belge akışları | Opsiyonel, fallback mevcut |

Uygulama interaktif modda eksik araçları kontrol eder ve kurulum için yönlendirir.

### Dönüşüm Yoluna Göre Zorunluluk

| Dönüşüm yolu | Öncelik sırası | Zorunlu araç |
|---|---|---|
| `md -> pdf` | `Pandoc -> LibreOffice -> Dahili Go renderer` | Zorunlu değil (harici araçsız da çalışır) |
| `html -> pdf` | `LibreOffice -> Dahili Go renderer` | Zorunlu değil (harici araçsız da çalışır) |
| `docx -> pdf` | `LibreOffice -> Metin tabanlı fallback` | Zorunlu değil (kalite için önerilir) |
| Ses/Video dönüşümleri | `FFmpeg` | Evet (`FFmpeg` zorunlu) |
| `csv -> xlsx` | `LibreOffice` | Evet (`LibreOffice` zorunlu) |
| Bazı `odt/rtf` hedefli ofis dönüşümleri | `LibreOffice` | Çoğu akışta zorunlu |

Notlar:
- `Windows`, `macOS` ve `Linux` üzerinde temel kurulum çalışır.
- Harici araçlar kurulmazsa sadece ilgili dönüşüm akışları etkilenir; tüm uygulama devre dışı kalmaz.
- En yüksek belge çıktı kalitesi için `md -> pdf` tarafında `Pandoc` (ve uygun PDF engine), ofis belgelerinde ise `LibreOffice` önerilir.

## Yapılandırma

- Konfigürasyon dosyası: `~/.docufy/config.json`
- Bu dosyada ilk çalıştırma bilgisi ve varsayılan çıktı dizini tutulur.
- İnteraktif moddan varsayılan çıktı dizinini değiştirebilirsiniz.
- Kullanıcı tanımlı profiller `~/.docufy/profiles/*.toml` altında tutulur.

### Proje bazlı yapılandırma (`.docufy.toml`)

CLI, çalışma dizininden başlayıp üst dizinlere çıkarak `.docufy.toml` arar.
Hazır örnek için: `.docufy.toml.example` dosyasını kopyalayabilirsiniz.

Örnek:
```toml
default_output = "./output"
workers = 8
quality = 85
profile = "social-story"
on_conflict = "versioned"
metadata_mode = "strip"
retry = 2
retry_delay = "1s"
report_format = "json"
```

Öncelik sırası:
1. CLI flag
2. Environment variable
3. `.docufy.toml`
4. Uygulama varsayılanı

Desteklenen environment variable'lar:
- `DOCUFY_OUTPUT`
- `DOCUFY_WORKERS`
- `DOCUFY_QUALITY`
- `DOCUFY_PROFILE`
- `DOCUFY_ON_CONFLICT`
- `DOCUFY_METADATA`
- `DOCUFY_RETRY`
- `DOCUFY_RETRY_DELAY`
- `DOCUFY_REPORT`

## Sorun Giderme

### `command not found: docufy`
- `PATH` içine `$(go env GOPATH)/bin` ekleyin.
- Terminali yeniden açın.

### Eski sürüm/eskimiş help çıktısı görünüyor
```bash
cd /proje/dizini
go install ./cmd/docufy
which docufy
docufy --help
```

### Dönüşüm desteklenmiyor hatası
Önce formatları doğrulayın:
```bash
docufy formats --from <kaynak>
docufy formats --to <hedef>
```

### FFmpeg bulunamadı
macOS:
```bash
brew install ffmpeg
```
Linux (Debian/Ubuntu):
```bash
sudo apt install ffmpeg
```

## Release
GitHub'a `vX.Y.Z` formatında bir tag push edildiğinde [`.github/workflows/release.yml`](.github/workflows/release.yml) otomatik olarak:

- `go test ./...` çalıştırır
- macOS, Linux ve Windows için `amd64` + `arm64` binary arşivleri üretir
- `checksums.txt` oluşturur ve tüm artefact'ları GitHub Release'e yükler

## Geliştirme
```bash
git clone https://github.com/mlihgenel/docufy.git
cd docufy
go test ./...
go run . --help
```

## Proje Yapısı
```text
docufy/
├── cmd/docufy/ # Uygulama giriş noktası (main package)
├── docs/assets/          # README ve proje dokümantasyon görselleri
├── examples/             # Örnek pipeline ve kullanım dosyaları
├── internal/cli/         # Cobra komutları ve TUI akışları
├── internal/converter/   # Dönüştürme motorları (document, image, audio, video)
├── internal/batch/       # Worker pool ve batch yürütme
├── internal/pipeline/    # Çok adımlı pipeline yürütme
├── internal/watch/       # Klasör izleme altyapısı
├── internal/profile/     # Built-in ve kullanıcı profilleri
├── internal/config/      # Uygulama ayarları
├── internal/installer/   # Bağımlılık kontrol/kurulum yardımcıları
└── internal/ui/          # Ortak terminal UI yardımcıları
```

## Katkı
Katkılar memnuniyetle karşılanır.

1. Repo'yu fork edin.
2. Yeni branch açın.
3. Değişiklikleri yapın.
4. Testleri çalıştırın.
5. Pull request gönderin.

Issue ve öneriler için: [GitHub Issues](https://github.com/mlihgenel/docufy/issues)

## Lisans
Bu proje [MIT Lisansı](LICENSE) ile lisanslanmıştır.
