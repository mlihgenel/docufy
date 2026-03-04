# AI Mode + Strands Entegrasyon Mimarisi

> Tarih: 2026-03-04  
> Proje: Docufy (`github.com/mlihgenel/docufy/v2`)

## 1) Hedef

Uygulamada iki kullanım modunu birlikte sunmak:

1. **Manuel Mod**: mevcut CLI/TUI akışları (convert, batch, watch, video trim, extract-audio vb.)
2. **AI Modu**: kullanıcı doğal dil ile komut verir, AI arka planda aynı işlemleri tool çağrıları ile çalıştırır.

Örnek AI komutları:
- "Bu dosyayı PNG'ye dönüştür."
- "Bu videonun 20 ile 30. saniyesini klip olarak kırp."
- "Bu toplantı videosundan metin çıkar ve özetle."

## 2) Neden Bu Tasarım?

Kod tabanı Go (Cobra + Bubble Tea). Strands ekosisteminde resmi ve olgun SDK tarafı Python/TypeScript yönünde.  
Bu nedenle önerilen mimari:

- **Go uygulama**: UI, dönüşüm engine, dosya işlemleri, validasyon, güvenli tool execution.
- **Strands sidecar** (öneri: Python): agent orkestrasyonu ve doğal dil planlama.

Bu ayrım ile:
- TUI tek süreçte stabil kalır.
- Strands entegrasyonu bağımsız geliştirilir.
- İleride farklı agent/provider değişimleri Go çekirdeğini bozmaz.

## 3) Yüksek Seviye Mimarî

```text
Kullanıcı (TUI)
   |
   v
Go TUI (AI ekranı)
   |
   | local IPC (HTTP localhost veya stdio JSON-RPC)
   v
Strands Sidecar (Planner/Executor/Verifier)
   |
   | tool çağrıları (sadece izinli)
   v
Go Tool Gateway (convert/trim/extract/transcribe wrappers)
   |
   v
Mevcut converter/pipeline/video komutları
```

## 4) Çoklu Agent Kullanımı (Bu Projeye Uygun)

Bu projede çoklu-agent **mümkün** ve **anlamlı**:

- **Planner Agent**
  - Kullanıcı niyetini parse eder.
  - Gerekli adımları planlar.
  - Riskli aksiyonları "onay gerekiyor" olarak işaretler.

- **Executor Agent**
  - Planı tool çağrılarına dönüştürür.
  - Her adımda structured argümanlarla Go tarafını çağırır.

- **Verifier Agent**
  - Üretilen çıktıyı kontrol eder (dosya var mı, süre/format doğru mu).
  - Gerekirse retry/alternatif codec planı üretir.

Opsiyonel:
- **Transcription/Summary Agent**
  - Ses->metin (provider/local) sonrası özet, action item, başlık üretir.

> Not: Swarm deseni de kullanılabilir; ancak ilk sürüm için deterministik işlerde Graph/Workflow yaklaşımı daha kontrollüdür.

## 5) Tool Sözleşmeleri (AI'nin Çağıracağı İzinli Fonksiyonlar)

AI hiçbir zaman rastgele shell komutu üretmez. Sadece aşağıdaki tool'ları çağırır:

1. `convert_file`
   - input: `input_path`, `to`, `quality?`, `output_dir?`, `name?`, `metadata_mode?`
   - output: `status`, `output_path`, `duration_ms`, `error?`

2. `trim_video`
   - input: `input_path`, `mode(clip|remove)`, `start`, `end?`, `duration?`, `ranges?`, `codec?`, `to?`
   - output: `status`, `output_path`, `plan?`, `error?`

3. `extract_audio`
   - input: `input_path`, `to`, `quality?`, `copy?`, `name?`
   - output: `status`, `output_path`, `error?`

4. `get_file_info`
   - input: `path`
   - output: mevcut `info` çıktısına benzer metadata

5. `transcribe_media` (Phase 2)
   - input: `input_path`, `provider`, `model?`, `language?`, `diarization?`
   - output: `transcript_path`, `segments?`, `error?`

6. `summarize_transcript` (Phase 2)
   - input: `transcript_path`, `style?`, `target_language?`
   - output: `summary_path` veya `summary_text`

## 6) TUI Entegrasyonu

`interactive.go` içinde yeni AI alanı açılır:

- Yeni section:
  - `ID: "ai"`
  - `Label: "AI Asistan"`
  - `Desc: "Doğal dil ile işlemleri otomatik yaptır"`

- Yeni action:
  - `menuActionAIAssistant`

- Yeni state'ler:
  - `stateAIIntro`
  - `stateAIAuthProviderSelect`
  - `stateAIAuthInput`
  - `stateAIChat`
  - `stateAIPlanConfirm`
  - `stateAIExecuting`
  - `stateAIDone`

- Yeni model alanları:
  - `aiEnabled bool`
  - `aiProvider string`
  - `aiModel string`
  - `aiSessionID string`
  - `aiMessages []chatMessage`
  - `aiPendingPlan *plan`
  - `aiLastResult string`
  - `aiError string`

## 7) Kimlik Doğrulama ve Secret Yönetimi

Mevcut `internal/config/config.go` düz JSON saklıyor; API key burada plaintext tutulmamalı.

Öneri:
- Config'te sadece non-secret ayarlar:
  - provider, model, base_url, default_behavior
- Secret için OS keychain:
  - macOS Keychain / Windows Credential Manager / Linux Secret Service
  - Go tarafında `go-keyring` benzeri paket ile erişim

İlk AI giriş akışı:
1. Provider seç.
2. API key gerekiyorsa tek sefer gir.
3. Keychain'e kaydet.
4. Test ping başarılıysa AI chat ekranına geç.

## 8) Güvenlik Kuralları

1. AI sadece kayıtlı tool listesini çağırır.
2. Path erişimi whitelist ile sınırlandırılır.
3. Riskli aksiyonlarda kullanıcı onayı zorunludur:
   - overwrite
   - remove mode trim
   - batch/watch toplu işlemler
4. Her tool çağrısı audit log olarak kaydedilir (timestamp + args + result).
5. Prompt injection etkisini azaltmak için:
   - Tool argümanları şema doğrulamasından geçer.
   - Serbest metin doğrudan shell'e verilmez.

## 9) Failover ve Dayanıklılık

- Sidecar çalışmıyorsa:
  - TUI AI ekranında "AI servis başlatılamadı" + manuel moda geri dönüş.
- Tool timeout:
  - iş iptal/yeniden dene seçeneği.
- Uzun işlemler:
  - mevcut FFmpeg progress callback'leri AI modunda da gösterilir.

## 10) Uygulama Fazları

### Faz 1 (MVP): AI Kontrollü Dönüşüm
- AI ekranı + provider setup + chat
- Tool set: `convert_file`, `trim_video`, `extract_audio`, `get_file_info`
- Planner + Executor + Verifier (temel)

### Faz 2: Ses->Metin + Özet
- `transcribe_media`
- `summarize_transcript`
- transcript dosyası + markdown özet çıktısı

### Faz 3: Pipeline ve Watch AI
- AI planını pipeline JSON'a dökebilme
- "Bu klasörü izle ve gelen videoları ... yap" gibi watch görevleri

## 11) Dosya Bazlı Uygulama Planı

1. `internal/cli/interactive.go`
   - AI section/action/state/model alanları ekleme
   - AI ekran render ve input handler

2. `internal/config/config.go`
   - AI non-secret ayarlar (`AIProvider`, `AIModel`, `AIBaseURL`) ekleme

3. `internal/ai/`
   - `types.go` (tool request/response modelleri)
   - `gateway.go` (Go tool wrapper katmanı)
   - `client.go` (sidecar iletişim katmanı)
   - `policy.go` (onay/güvenlik kuralları)

4. `cmd/docufy/main.go` ve root init akışı
   - sidecar lifecycle (opsiyonel otomatik başlat)

5. `docs/`
   - kullanım ve güvenlik dokümanı

## 12) Test Stratejisi

- Unit:
  - tool arg doğrulama
  - riskli işlem policy'si
  - config serileştirme

- Integration:
  - sahte sidecar ile AI plan -> tool execution akışı
  - trim/convert gerçek dosya fixture testleri

- TUI:
  - AI state geçişleri
  - auth->chat->confirm->execute akışı

## 13) Başarı Kriteri

Kullanıcı tek bir komutla şunları yapabiliyorsa MVP başarılıdır:

> "Bu videonun 20-30 saniyesini kırp, mp4 olarak ver ve metadata'yı temizle."

Beklenen:
1. AI planı gösterilir.
2. Kullanıcı onaylar.
3. Tool çağrıları çalışır.
4. Çıktı dosyası üretilir ve TUI'da raporlanır.
