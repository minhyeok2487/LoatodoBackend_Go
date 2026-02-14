# Spring -> Go Migration API 테스트 결과

## 테스트 환경
- Spring: localhost:8888
- Go: localhost:8889
- 테스트 계정: authtest2@test.com (38개 캐릭터)
- 테스트 캐릭터: 이다 (ID: 191741, 아이템레벨: 1790.83)

## Character List API (3개) ✅ 완료

| # | API | 결과 | 비고 |
|---|-----|------|------|
| 1 | GET /api/v1/character-list | ✅ 동일 | 99691 bytes |
| 2 | GET /api/v1/character-list/deleted | ✅ 동일 | 빈 배열 [] |
| 3 | PATCH /api/v1/character-list/sorting | ✅ 수정 후 동일 | 캐릭터 리스트 반환하도록 수정 |

## Character Basic API (7개) ✅ 완료

| # | API | 결과 | 비고 |
|---|-----|------|------|
| 4 | PATCH /api/v1/character/settings | ✅ 동일 | CharacterResponse 반환, chaos/guardian 콘텐츠 정보 추가 |
| 5 | PATCH /api/v1/character/gold-character | ✅ 동일 | CharacterResponse 반환 |
| 6 | POST /api/v1/character/memo | ✅ 동일 | CharacterResponse 반환 |
| 7 | PATCH /api/v1/character/deleted | ✅ 동일 | 빈 200 OK 반환 |
| 8 | PUT /api/v1/character | ⏳ 테스트 예정 | API Key 필요 |
| 9 | PATCH /api/v1/character/name | ✅ 동일 | CharacterResponse 반환 |
| 10 | POST /api/v1/character | ⏳ 테스트 예정 | API Key 필요 |

## Character Day API (4개) ⏳ 대기

| # | API | 결과 | 비고 |
|---|-----|------|------|
| 11 | PATCH /api/v1/character/day-check | ⏳ | |
| 12 | POST /api/v1/character/day-check/all | ⏳ | |
| 13 | POST /api/v1/character/day-gauge | ⏳ | |
| 14 | POST /api/v1/character/day-rest-all | ⏳ | |

## Character Week API (14개) ⏳ 대기

| # | API | 결과 | 비고 |
|---|-----|------|------|
| 15-28 | 주간 컨텐츠 API | ⏳ | |

## 수정 사항 기록

### 2026-02-14

1. **UpdateSorting 응답 수정** (commit e06b5fa)
   - 기존: `{"message":"ok"}` 반환
   - 수정: 업데이트된 캐릭터 리스트 반환

2. **Character API 응답 수정** (이번 커밋)
   - UpdateSettings, ToggleGoldCharacter, UpdateMemo, ChangeCharacterName: CharacterResponse 반환
   - ToggleDeleted: 빈 200 OK 반환 (Spring과 동일)
   - GetCharacterByID 함수 추가 (단일 캐릭터 조회)
   - getDayContent 함수 추가 (chaos/guardian 콘텐츠 정보 조회)

## 메모 (확인 필요 사항)
- 숫자 형식 차이: Spring은 `1720.0`, Go는 `1720` (의미적으로 동일)
- JSON 필드 순서 차이 (의미적으로 동일)
