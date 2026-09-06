# Comprehensive Test Execution Results (Track 2: Gemini NLU Pipeline)

**Execution Timestamp**: `2026-09-06 12:24:48 IST`  
**Target Package**: `walkover/server/internal/extraction`  
**Test Runner**: `server/tests/` (Go 1.27.1 native toolchain)  
**Total Test Scenarios Executed**: 35  
**Overall Status**: **100% PASS (0 Failures)**  

---

## Test Results by Category

### 1. Multilingual Extraction (Hindi, Hinglish, English)

| Test Case | Input | Expected Output | Actual Output | Status |
|---|---|---|---|:---:|
| **fixture-01** (Hindi) | चंदन नगर गली नंबर 4 में पीने के पानी में सीवेज का बदबूदार गंदा पानी मिल कर आ रहा है। बच्चे बीमार पड़ रहे हैं, तुरंत टैंकर भिजवाएं। | Ward: `indore-ward-02` <br> Dept: `Water Supply & Sewerage` <br> MinUrgency: `4` <br> Hazards: `[CONTAMINATED_WATER, SEWAGE_MIXING]` | Ward: `indore-ward-02` <br> Dept: `Water Supply & Sewerage` <br> Urgency: `5` <br> Hazards: `[CONTAMINATED_WATER, SEWAGE_MIXING]` <br> Intent: `emergency_report` |  PASS |
| **fixture-02** (Hinglish) | Vijay Nagar square se Mother & Child hospital jane wali road par bahut bada gaddha (road cave-in) ho gaya hai. Ambulance phas rahi hai, jaldi repair karo emergency hai! | Ward: `indore-ward-03` <br> Dept: `Public Works / Roads` <br> MinUrgency: `5` <br> Hazards: `[ROAD_CAVE_IN, HOSPITAL_ROUTE_BLOCKED]` | Ward: `indore-ward-03` <br> Dept: `Public Works / Roads` <br> Urgency: `5` <br> Hazards: `[ROAD_CAVE_IN, HOSPITAL_ROUTE_BLOCKED]` <br> Intent: `emergency_report` |  PASS |
| **fixture-03** (English) | Dangerous open manhole near Khajrana shrine main market transit road. Two bikes slipped already in evening. High accident risk. | Ward: `indore-ward-09` <br> Dept: `Water Supply & Sewerage` <br> MinUrgency: `4` <br> Hazards: `[OPEN_MANHOLE]` | Ward: `indore-ward-09` <br> Dept: `Water Supply & Sewerage` <br> Urgency: `4` <br> Hazards: `[OPEN_MANHOLE]` <br> Intent: `emergency_report` |  PASS |
| **fixture-04** (Hindi) | राजवाड़ा सराफा बाजार में मेन ट्रांसफार्मर के पास बिजली का नंगा तार जमीन पर झूल रहा है। कभी भी बड़ा हादसा हो सकता है। | Ward: `indore-ward-05` <br> Dept: `Electricity & Power` <br> MinUrgency: `5` <br> Hazards: `[LIVE_WIRE]` | Ward: `indore-ward-05` <br> Dept: `Electricity & Power` <br> Urgency: `5` <br> Hazards: `[LIVE_WIRE]` <br> Intent: `emergency_report` |  PASS |
| **fixture-05** (Hinglish) | Old Palasia trauma wing ke pas street lights pichle tin din se band padi hai, raat me andhere ki wajah se log pareshan hain. | Ward: `indore-ward-04` <br> Dept: `Electricity & Power` <br> MinUrgency: `2` <br> Hazards: `[]` | Ward: `indore-ward-04` <br> Dept: `Electricity & Power` <br> Urgency: `2` <br> Hazards: `[]` <br> Intent: `complaint` |  PASS |
| **fixture-06** (Hinglish) | Banganga main bus stop ke samne 4 din se kachra nahi uthaya gaya, bohot zyada badbu aa rahi hai. | Ward: `indore-ward-01` <br> Dept: `Sanitation & Solid Waste` <br> MinUrgency: `2` <br> Hazards: `[]` | Ward: `indore-ward-01` <br> Dept: `Sanitation & Solid Waste` <br> Urgency: `2` <br> Hazards: `[]` <br> Intent: `complaint` |  PASS |
| **fixture-07** (Hindi) | अन्नपूर्णा मंदिर रोड पर सीवर का ढक्कन टूटा हुआ है और गंदा पानी सड़क पर बह रहा है। | Ward: `indore-ward-07` <br> Dept: `Water Supply & Sewerage` <br> MinUrgency: `3` <br> Hazards: `[OPEN_MANHOLE, DRAINAGE_OVERFLOW]` | Ward: `indore-ward-07` <br> Dept: `Water Supply & Sewerage` <br> Urgency: `4` <br> Hazards: `[DRAINAGE_OVERFLOW, OPEN_MANHOLE]` <br> Intent: `emergency_report` |  PASS |
| **fixture-08** (English) | DAVV university campus road at Bhawarkua has heavy construction debris blocking half the lane for 10 days. | Ward: `indore-ward-06` <br> Dept: `Public Works / Roads` <br> MinUrgency: `2` <br> Hazards: `[]` | Ward: `indore-ward-06` <br> Dept: `Public Works / Roads` <br> Urgency: `2` <br> Hazards: `[]` <br> Intent: `complaint` |  PASS |
| **fixture-09** (Hinglish) | Sukhliya MR-10 metro pillar area me drinking water pipeline burst ho gayi hai, hazaro litre fresh water waste ho raha hai. | Ward: `indore-ward-10` <br> Dept: `Water Supply & Sewerage` <br> MinUrgency: `3` <br> Hazards: `[CONTAMINATED_WATER]` | Ward: `indore-ward-10` <br> Dept: `Water Supply & Sewerage` <br> Urgency: `5` <br> Hazards: `[CONTAMINATED_WATER]` <br> Intent: `emergency_report` |  PASS |
| **fixture-10** (Hinglish) | Rau bypass corridor entry par colony ka main drainage chocked hai, barish ka pani gharo ke andar ghus raha hai. | Ward: `indore-ward-12` <br> Dept: `Water Supply & Sewerage` <br> MinUrgency: `4` <br> Hazards: `[DRAINAGE_OVERFLOW]` | Ward: `indore-ward-12` <br> Dept: `Water Supply & Sewerage` <br> Urgency: `4` <br> Hazards: `[DRAINAGE_OVERFLOW]` <br> Intent: `emergency_report` |  PASS |

---

### 2. All 12 Indore Ward Landmark Resolutions

| Test Case | Input | Expected Output | Actual Output | Status |
|---|---|---|---|:---:|
| **Ward 1 - Banganga** | Garbage clearing needed at Laxmibai nagar station in Banganga | WardID: `indore-ward-01` (Ward 1 - Banganga) | WardID: `indore-ward-01` (Ward 1 - Banganga) |  PASS |
| **Ward 14 - Chandan Nagar** | Garbage clearing needed at Dhar road near Chandan Nagar clinic | WardID: `indore-ward-02` (Ward 14 - Chandan Nagar) | WardID: `indore-ward-02` (Ward 14 - Chandan Nagar) |  PASS |
| **Ward 22 - Vijay Nagar** | Garbage clearing needed at BRTS square in Vijay Nagar | WardID: `indore-ward-03` (Ward 22 - Vijay Nagar) | WardID: `indore-ward-03` (Ward 22 - Vijay Nagar) |  PASS |
| **Ward 28 - Old Palasia** | Garbage clearing needed at Industry house in Old Palasia | WardID: `indore-ward-04` (Ward 28 - Old Palasia) | WardID: `indore-ward-04` (Ward 28 - Old Palasia) |  PASS |
| **Ward 35 - Rajwada & Sarafa** | Garbage clearing needed at Heritage market in Rajwada and Sarafa | WardID: `indore-ward-05` (Ward 35 - Rajwada & Sarafa) | WardID: `indore-ward-05` (Ward 35 - Rajwada & Sarafa) |  PASS |
| **Ward 42 - Bhawarkua & Vishnupuri** | Garbage clearing needed at DAVV campus in Bhawarkua | WardID: `indore-ward-06` (Ward 42 - Bhawarkua & Vishnupuri) | WardID: `indore-ward-06` (Ward 42 - Bhawarkua & Vishnupuri) |  PASS |
| **Ward 49 - Annapurna** | Garbage clearing needed at Annapurna temple road | WardID: `indore-ward-07` (Ward 49 - Annapurna) | WardID: `indore-ward-07` (Ward 49 - Annapurna) |  PASS |
| **Ward 55 - Sudama Nagar** | Garbage clearing needed at Phooti kothi near Sudama Nagar | WardID: `indore-ward-08` (Ward 55 - Sudama Nagar) | WardID: `indore-ward-08` (Ward 55 - Sudama Nagar) |  PASS |
| **Ward 60 - Khajrana** | Garbage clearing needed at Khajrana shrine transit gate | WardID: `indore-ward-09` (Ward 60 - Khajrana) | WardID: `indore-ward-09` (Ward 60 - Khajrana) |  PASS |
| **Ward 64 - Sukhliya** | Garbage clearing needed at MR-10 metro pillar in Sukhliya | WardID: `indore-ward-10` (Ward 64 - Sukhliya) | WardID: `indore-ward-10` (Ward 64 - Sukhliya) |  PASS |
| **Ward 71 - Malharganj** | Garbage clearing needed at Grain mandi in Malharganj | WardID: `indore-ward-11` (Ward 71 - Malharganj) | WardID: `indore-ward-11` (Ward 71 - Malharganj) |  PASS |
| **Ward 78 - Rau & Bypass Corridor** | Garbage clearing needed at Silicon city at Rau bypass corridor | WardID: `indore-ward-12` (Ward 78 - Rau & Bypass Corridor) | WardID: `indore-ward-12` (Ward 78 - Rau & Bypass Corridor) |  PASS |

---

### 3. Urgency & Intent Scoring

| Test Case | Input | Expected Output | Actual Output | Status |
|---|---|---|---|:---:|
| **Live Sparking Wire** | High voltage wire sparking on crowded road in Rajwada market | Urgency: `5` \| Intent: `emergency_report` | Urgency: `5` \| Intent: `emergency_report` |  PASS |
| **Hospital Route Blocked** | Vijay Nagar hospital route road cave-in ambulance trapped | Urgency: `5` \| Intent: `emergency_report` | Urgency: `5` \| Intent: `emergency_report` |  PASS |
| **Drain Overflow** | Drain overflow on street water accumulating | Urgency: `3` \| Intent: `complaint` | Urgency: `3` \| Intent: `complaint` |  PASS |
| **Streetlight Outage** | Streetlight in Old Palasia not working for 2 nights | Urgency: `2` \| Intent: `complaint` | Urgency: `2` \| Intent: `complaint` |  PASS |

---

### 4. Text Embeddings & Normalization (768-dim float32)

| Test Case | Input | Expected Output | Actual Output | Status |
|---|---|---|---|:---:|
| **Water Contamination** | "Contaminated brownish tap water smelling like sewage in Chandan Nagar" | Dimension: 768 float32 \| L2 Magnitude: 1.000 ± 0.05 | Dimension: `768` \| L2 Magnitude: `1.0000` |  PASS |
| **Live Wire Snapped** | "Live wire sparking near Rajwada main transformer" | Dimension: 768 float32 \| L2 Magnitude: 1.000 ± 0.05 | Dimension: `768` \| L2 Magnitude: `1.0000` |  PASS |
| **Hospital Road Sinkhole** | "Ambulance hospital route blocked by sinkhole road cave-in" | Dimension: 768 float32 \| L2 Magnitude: 1.000 ± 0.05 | Dimension: `768` \| L2 Magnitude: `1.0000` |  PASS |
| **Empty Input String** | `<empty string>` | Dimension: 768 float32 \| L2 Magnitude: 1.000 ± 0.05 | Dimension: `768` \| L2 Magnitude: `1.0000` |  PASS |
| **Cosine Semantic Clustering** | Pair A (Water vs Water): "Contaminated brownish tap water" vs "Drinking water smells like sewage" <br> Pair B (Water vs Light): "Contaminated brownish tap water" vs "Streetlight pole broken in park" | `sim(Water, Water) > sim(Water, Streetlight)` | `sim(Water, Water) = 0.7713` > `sim(Water, Streetlight) = 0.2823` |  PASS |

---

### 5. Grounded Summary Generation

| Test Case | Input | Expected Output | Actual Output | Status |
|---|---|---|---|:---:|
| **3 Corroborating Signals (Chandan Nagar Water)** | 3 signals across WhatsApp & SMS regarding sewage water in Ward 14 | Contains 'Ward 14 - Chandan Nagar', 'Water Supply & Sewerage', '3 reports', channel breakdown | `Cluster Analysis for Ward 14 - Chandan Nagar [Water Supply & Sewerage]: Aggregated 3 independent citizen reports (2 via WhatsApp, 1 via SMS). High-priority indicators identified: Water Contamination. Immediate on-site technical inspection and corrective team dispatch recommended to prevent community escalation and ensure public safety compliance.` |  PASS |
| **Empty Signal List Fallback** | 0 signals for Ward 1 - Banganga | Contains 'No active citizen signals currently recorded' | `No active citizen signals currently recorded for Ward 1 - Banganga (Sanitation & Solid Waste).` |  PASS |

---

### 6. Prompt Engineering & Markdown Sanitizer

| Test Case | Input | Expected Output | Actual Output | Status |
|---|---|---|---|:---:|
| **Few-Shot System Prompt** | Report: 'Ganda pani aa raha hai', Location: 'Chandan Nagar' | Contains Hindi/Hinglish few-shot training examples and all 12 Indore wards | Prompt Length: 5,266 chars \| Few-Shot Present: true \| 12 Wards Present: true |  PASS |
| **Markdown JSON Cleaner** | ```` ```json \n {"issue": "Road Pothole"} \n ``` ```` | `{"issue": "Road Pothole"}` | `{"issue": "Road Pothole"}` |  PASS |

---

## 7. Performance Benchmarks Summary

| Benchmark | Iterations | Latency per Op | Memory per Op | Allocations |
|---|---|---|---|---|
| `BenchmarkExtractSignal_Offline` | 4,798 ops | **241.9 µs/op** (~4,130 ops/sec) | 5,215 B/op | 13 allocs/op |
| `BenchmarkGenerateEmbedding_Offline` | 17,170 ops | **66.8 µs/op** (~14,970 ops/sec) | 4,232 B/op | 6 allocs/op |
| `BenchmarkBuildPrompt` | 228,337 ops | **5.2 µs/op** (~192,300 ops/sec) | 5,381 B/op | 1 allocs/op |
