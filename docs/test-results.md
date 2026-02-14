# Spring -> Go 백엔드 테스트 결과

## 테스트 환경
- **Go**: localhost:8889
- **Spring**: localhost:8888
- **DB**: MySQL 3307
- **테스트 일시**: 2026-02-14

---

## 1. N+1 쿼리 최적화 테스트 (완료)

### 결과 요약
| API | Go (req/s) | Spring (req/s) | 비율 |
|-----|------------|----------------|------|
| Friends API | 644.2 | 134.5 | **4.79x** |
| Character List | 690.4 | 388.4 | **1.78x** |
| Server Todos | 843.4 | 369.2 | **2.28x** |
| Custom Todo | 1029.3 | 529.3 | **1.94x** |
| Raid Category | 976.7 | 998.7 | 0.98x |

### 수정 사항
- `GetFriends`: N+1 역방향 쿼리 → 배치 쿼리로 변경
- `SearchCharacter`: N+1 친구 상태 조회 → LEFT JOIN으로 변경
- `GetComments`: N+1 사용자명 조회 → JOIN으로 변경
- `AddRemoveWeekRaid`: N+1 콘텐츠 조회 → 배치 쿼리로 변경

---

## 2. 쓰기 API 테스트 - 2026-02-14

### 기능 테스트 결과
| 테스트 | Go 상태 | Spring 상태 | 결과 |
|--------|---------|-------------|------|
| 일일숙제 체크 | 200 (9ms) | 200 (13ms) | ✅ PASS |
| 휴식게이지 업데이트 | 200 (7ms) | 200 (12ms) | ✅ PASS |
| 모든 일일숙제 체크 | 200 (8ms) | 200 (16ms) | ✅ PASS |
| 주간에포나 체크 | 200 (8ms) | 200 (12ms) | ✅ PASS |
| 실마엘 토글 | 200 (7ms) | 200 (12ms) | ✅ PASS |
| 큐브 티켓 업데이트 | 200 (7ms) | 200 (12ms) | ✅ PASS |
| 커스텀 할일 생성 | 201 (5ms) | 200 (18ms) | ✅ PASS |
| 캐릭터 검색 (없는 캐릭터) | 200 (54ms) | 400 (173ms) | ⚠️ 동작 차이 |
| 친구 목록 조회 | 200 (5ms) | 200 (12ms) | ✅ PASS |
| 레이드 카테고리 조회 | 200 (3ms) | 200 (4ms) | ✅ PASS |

**총 결과: 9/10 통과 (90.0%)**

### 부하 테스트 결과 (50 동시 요청)
| API | Go (req/s) | Spring (req/s) | 비율 |
|-----|------------|----------------|------|
| 일일숙제 체크 | 496.9 | 292.6 | **1.70x** |
| 커스텀 할일 생성 | 805.4 | 481.6 | **1.67x** |

### 발견된 동작 차이
1. **캐릭터 검색 (없는 캐릭터)**
   - Go: 200 OK + 빈 배열 `[]` 반환
   - Spring: 400 Bad Request + 에러 메시지 반환
   - **참고**: Go의 동작이 REST API 관점에서 더 적절함 (검색 결과 없음 = 빈 결과)

### 테스트된 API 목록
- POST `/api/v1/character/day/check` - 일일숙제 체크
- POST `/api/v1/character/day/gauge` - 휴식게이지 업데이트
- POST `/api/v1/character/day/check/all` - 모든 일일숙제 체크
- POST `/api/v1/character/week/epona` - 주간에포나 체크
- POST `/api/v1/character/week/silmael` - 실마엘 토글
- POST `/api/v1/character/week/cube` - 큐브 티켓 업데이트
- POST `/api/v1/custom` - 커스텀 할일 생성
- GET `/api/v1/friend/character/{name}` - 캐릭터 검색
- GET `/api/v1/friend` - 친구 목록 조회
- GET `/api/v1/schedule/raid/category` - 레이드 카테고리 조회

---

## 테스트 스크립트
- `/tmp/write_api_test.py` - 쓰기 API 테스트
- `/tmp/multi_api_load_test_v2.py` - 다중 API 부하 테스트
