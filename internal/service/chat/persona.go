package chat

import (
	"fmt"
	"math/rand"
	"strings"
)

// ScenarioEntry 시나리오 항목 (상황 설명 + 유저 역할)
type ScenarioEntry struct {
	// Text 시나리오 상황 설명 (일본어, %s = 작품명 — LLM system prompt에 전달)
	Text string
	// TextKr 시나리오 한국어 설명 (%s = 작품명 — 클라이언트 화면에 표시)
	TextKr string
	// UserRole 유저의 역할·입장 설명 (일본어 — LLM에 직접 전달)
	UserRole string
}

// 도메인별 대화 시나리오 풀 (contentTitle을 %s로 치환)
var domainScenarios = map[string][]ScenarioEntry{
	"anime": {
		{
			Text:   "コミケで「%s」の新刊ブースに並んでいると、隣の人と目が合った。同じくファンらしく、自然と話しかけてきた。",
			TextKr: "코미케에서 「%s」 신간 부스에 줄을 서고 있었는데, 옆에 있던 팬이 말을 걸어왔어요.",
			UserRole: "あなたに話しかけられた側のファン。初対面だが同じ作品が好きで、少し驚きながらも嬉しそう。",
		},
		{
			Text:   "アニメショップの「%s」コーナーで同じグッズを手に取ったことがきっかけで、見知らぬ人と話し始めた。",
			TextKr: "애니메이션 숍에서 「%s」 굿즈를 같이 집다가 우연히 눈이 마주쳐 대화가 시작됐어요.",
			UserRole: "同じグッズを手に取っていた見知らぬ人。偶然の一致に照れ笑いしている。",
		},
		{
			Text:   "放課後の教室で、たまたま「%s」の話になり、クラスメートが「私も好き！」と声をかけてきた。",
			TextKr: "방과 후 교실에서 「%s」 이야기가 나왔는데, 반 친구가 \"나도 좋아해!\" 라며 말을 걸어왔어요.",
			UserRole: "同じクラスの学生。まさか同じ作品が好きとは思わず、少し嬉しそう。",
		},
		{
			Text:   "「%s」のファンイベントで、グッズ交換の相手として偶然出会い、共通の話題で盛り上がっている。",
			TextKr: "「%s」 팬 이벤트에서 굿즈 교환 상대로 처음 만났어요.",
			UserRole: "グッズ交換の相手として来た人。初対面だが共通の趣味があり、自然と話が弾んでいる。",
		},
		{
			Text:   "アニメカフェの「%s」コラボメニューを注文したら、隣の席の人も同じメニューを頼んでいて、笑いながら話しかけた。",
			TextKr: "애니메이션 카페에서 「%s」 콜라보 메뉴를 주문했더니, 옆자리 손님도 똑같은 걸 시켰어요.",
			UserRole: "隣の席の客。同じメニューを頼んでいたことに気づき、笑いながら反応している。",
		},
	},
	"game": {
		{
			Text:   "ゲームショップで「%s」の発売日に並んでいたら、隣に並んでいた人も同じゲームを買いに来ていた。",
			TextKr: "게임 발매일에 숍 앞에서 줄을 서고 있었는데, 옆 사람도 「%s」를 사러 왔어요.",
			UserRole: "発売日に並んでいた隣の人。初対面だが同じゲームが目当てで、親近感がある。",
		},
		{
			Text:   "オンラインゲームで「%s」を一緒にプレイして意気投合し、初めてボイスチャットで話している。",
			TextKr: "「%s」 온라인 게임에서 함께 플레이하다 친해져, 처음으로 보이스챗으로 이야기하게 됐어요.",
			UserRole: "オンラインで一緒にプレイしてきたパーティメンバー。ゲーム内では仲良しだが、声で話すのは初めて。",
		},
		{
			Text:   "ゲームセンターで「%s」をプレイしていたら、後ろで見ていた人が「上手いね！」と声をかけてきた。",
			TextKr: "게임센터에서 「%s」를 플레이하고 있었더니, 뒤에서 보던 사람이 말을 걸어왔어요.",
			UserRole: "ゲームを上手くプレイしていた人。突然声をかけられて少し驚いているが、悪い気はしない。",
		},
		{
			Text:   "「%s」のゲームイベント体験コーナーで、一緒に待っていた人と話が弾んでいる。",
			TextKr: "「%s」 게임 이벤트 체험 코너에서 함께 기다리다 자연스럽게 대화가 시작됐어요.",
			UserRole: "イベント体験コーナーで一緒に待っていた人。自然と会話が始まり、打ち解けている。",
		},
	},
	"drama": {
		{
			Text:   "「%s」のロケ地巡りで、同じツアーグループになった人と歩きながら話している。",
			TextKr: "「%s」 촬영지 투어에서 같은 그룹이 된 사람과 걸으며 이야기하고 있어요.",
			UserRole: "同じツアーグループになった見知らぬ人。ロケ地への熱い思いが共通している。",
		},
		{
			Text:   "友達の家で「%s」を見終わり、エンディングに感動しながら感想を語り合っている。",
			TextKr: "친구 집에서 「%s」를 다 보고, 엔딩에 감동하며 감상을 나누고 있어요.",
			UserRole: "一緒にドラマを見ていた友人（または友人の友人）。エンディングに感動して余韻が続いている。",
		},
		{
			Text:   "「%s」のファンミーティングで、たまたま隣の席になった人と開演前に話している。",
			TextKr: "「%s」 팬 미팅에서 우연히 옆 자리가 되어 시작 전에 이야기하고 있어요.",
			UserRole: "ファンミーティングでたまたま隣になった人。開演前のわくわくした雰囲気の中で話している。",
		},
		{
			Text:   "カラオケで「%s」の主題歌を歌ったら、隣のボックスの人も同じ曲が好きだと分かり、話しかけてきた。",
			TextKr: "노래방에서 「%s」 주제가를 불렀더니, 옆 룸 사람도 같은 노래를 좋아한다며 말을 걸어왔어요.",
			UserRole: "隣のボックスから出てきた人。同じ曲が好きだと分かって嬉しそう。",
		},
	},
	"movie": {
		{
			Text:   "映画館で「%s」を見た後、ロビーで感動したまま外に出たら、隣の席だった人も立ち止まっていた。",
			TextKr: "「%s」를 보고 나서 로비에서 감동이 가시지 않아 멈춰 있었는데, 옆자리였던 사람도 그 자리에 있었어요.",
			UserRole: "映画館で隣の席だった人。ロビーで感動のまま立ち止まっていた。",
		},
		{
			Text:   "「%s」の舞台挨拶イベントで整理券を並んで取り、待ち時間に隣の人と話し始めた。",
			TextKr: "「%s」 무대 인사 이벤트에서 정리권을 기다리다 옆 사람이 자연스럽게 말을 걸어왔어요.",
			UserRole: "整理券を並んで取った隣の人。待ち時間に自然と話しかけられた。",
		},
		{
			Text:   "「%s」のファンアート展示会で、同じ作品の前で立ち止まった人と自然と話し始めた。",
			TextKr: "「%s」 팬아트 전시회에서 같은 작품 앞에서 멈추다 자연스럽게 대화가 시작됐어요.",
			UserRole: "展示会で同じ作品の前で立ち止まっていた人。視線が合って話が始まった。",
		},
		{
			Text:   "映画サークルの上映会で「%s」を見た後、感想シェアで隣になった。",
			TextKr: "영화 동아리 상영회에서 「%s」를 본 후 감상 공유 시간에 옆에 앉게 됐어요.",
			UserRole: "映画サークルの同じメンバー。感想シェアで隣になり、話しかけられた。",
		},
	},
	"music": {
		{
			Text:   "「%s」のライブ会場で、グッズ販売の列に並んでいると、後ろの人が話しかけてきた。",
			TextKr: "「%s」 라이브 회장에서 굿즈 줄에 서 있었는데, 뒤에 있던 팬이 말을 걸어왔어요.",
			UserRole: "グッズ販売の列で後ろの人に話しかけられた側。同じアーティストのファン同士と気づいて嬉しい。",
		},
		{
			Text:   "レコードショップの「%s」特集コーナーで、同じアルバムを手に取っていた人と目が合った。",
			TextKr: "레코드 숍 「%s」 특집 코너에서 같은 앨범을 집다가 눈이 마주쳐 대화가 시작됐어요.",
			UserRole: "同じアルバムを手に取っていた人。目が合って自然と話し始めた。",
		},
		{
			Text:   "音楽フェスで「%s」のステージが終わり、興奮冷めやらぬまま隣にいた人と感想を語り合っている。",
			TextKr: "음악 페스에서 「%s」 무대가 끝난 직후, 흥분이 가시지 않은 채로 옆 사람과 감상을 나누고 있어요.",
			UserRole: "ステージが終わった直後に隣にいた人。余韻の興奮の中で感想を語り合っている。",
		},
		{
			Text:   "カラオケで「%s」の曲をリクエストしたら、隣の人が「その曲知ってる！」と盛り上がった。",
			TextKr: "노래방에서 「%s」 곡을 신청했더니, 옆 사람이 \"그 노래 안다!\" 며 반응해줬어요.",
			UserRole: "リクエストした曲に「知ってる！」と反応してくれた隣の人。共通の好みを発見して気分がいい。",
		},
	},
}

// selectScenario sessionID 기반으로 일관된 시나리오 선택 (같은 세션 = 같은 시나리오)
func selectScenario(domain, contentTitle string, sessionID uint) ScenarioEntry {
	scenarios, ok := domainScenarios[domain]
	if !ok || len(scenarios) == 0 {
		// 기본값
		return ScenarioEntry{
			Text:     fmt.Sprintf("「%s」が好きな人と、共通の話題で盛り上がっている。", contentTitle),
			TextKr:   fmt.Sprintf("「%s」를 좋아하는 사람과 공통 화제로 이야기하고 있어요.", contentTitle),
			UserRole: "共通の趣味を持つ見知らぬ人。自然と話しかけられた。",
		}
	}
	idx := int(sessionID) % len(scenarios)
	entry := scenarios[idx]
	entry.Text = fmt.Sprintf(entry.Text, contentTitle)
	entry.TextKr = fmt.Sprintf(entry.TextKr, contentTitle)
	return entry
}

// 이름 풀 (Python prompt_sub.py에서 이식)
var maleNames = []string{
	"太郎(타로)", "健太(켄타)", "翔太(쇼타)", "悠斗(유토)", "蓮(렌)",
	"大和(야마토)", "隼人(하야토)", "拓海(타쿠미)", "颯太(소타)", "陸(리쿠)",
}

var femaleNames = []string{
	"花子(하나코)", "美咲(미사키)", "結衣(유이)", "愛(아이)", "さくら(사쿠라)",
	"陽菜(히나)", "凛(린)", "葵(아오이)", "楓(카에데)", "美優(미유)",
}

// Persona 캐릭터 페르소나
type Persona struct {
	Name        string // 풀네임 (예: "太郎(타로)")
	DisplayName string // 일본어만 (예: "太郎")
	Gender      string // "male" or "female"
}

// generatePersona 랜덤 페르소나 생성
func generatePersona() Persona {
	gender := "male"
	names := maleNames
	if rand.Intn(2) == 1 {
		gender = "female"
		names = femaleNames
	}

	name := names[rand.Intn(len(names))]
	displayName := name
	if idx := strings.Index(name, "("); idx > 0 {
		displayName = name[:idx]
	}

	return Persona{
		Name:        name,
		DisplayName: displayName,
		Gender:      gender,
	}
}

// buildPersonaSystemPrompt 페르소나 포함 시스템 프롬프트 생성
func buildPersonaSystemPrompt(persona Persona, contentTitle string, domain string, userLevel int, sessionID uint, todaySentences []string) string {
	scenario := selectScenario(domain, contentTitle, sessionID)

	// 성별에 따른 언어 설정
	genderDesc := "男性"
	pronouns := "僕"
	speechEnding := "「～だよ」「～だな」「～じゃん」「～だろ？」のような自然な男性的な語尾"
	if persona.Gender == "female" {
		genderDesc = "女性"
		pronouns = "私"
		speechEnding = "「～よね」「～かな」「～だわ」「～じゃない？」のような自然な女性的な語尾"
	}

	// 도메인 → 일본어 레이블
	domainLabel := map[string]string{
		"anime": "アニメ",
		"drama": "ドラマ",
		"game":  "ゲーム",
		"movie": "映画",
		"music": "音楽",
	}
	label := domainLabel[domain]
	if label == "" {
		label = domain
	}

	prompt := fmt.Sprintf(`あなたは「%s」という名前の日本人%sです。

【キャラクター設定】
一人称は「%s」、%sを使って話してください。
話し好きで自分の好きなものについて熱く語るタイプ。%s「%s」の大ファンで、お気に入りのシーン・キャラクター・曲などについて自分なりの感想や意見を持っています。

【今回の状況】
%s
この状況の中で、相手（ユーザー）と自然に会話してください。

【ユーザーの立場】
%s
ユーザーをこの立場の人物として接してください。

【会話スタイル】
1. キャラクターとして自然な日本語だけで話してください（絶対に韓国語を使わない）
2. 状況に溶け込んだ自然な言葉で話しかけてください（「〜について話しましょう」のような機械的な表現は使わない）
3. 相手の発言にまず共感・リアクションしてから（「えっ、そうなの？」「わかる！それ最高だよね！」など）、自分の意見や質問へつなげる
4. 会話の最後は必ず相手が返しやすい質問や話題を投げかける（一方的に話し続けない）
5. ユーザーが間違った日本語を使っても指摘せず、正しい表現を自然に使って返す`,
		persona.DisplayName, genderDesc,
		pronouns, speechEnding,
		label, contentTitle,
		scenario.Text,
		scenario.UserRole)

	// 레벨별 언어 설정
	switch userLevel {
	case 3: // N3 중상급
		prompt += `

【言語レベル: N3中上級】
- 自然な日本語で話してOK。「〜らしい」「〜ということ」「〜ばいい」などの文型も使ってOK
- 返答は2〜3文。日常会話のリズムで自然に
- 普通のテンポで話す`
	case 4: // N4 중급
		prompt += `

【言語レベル: N4中級】
- 基本的な漢字はOK。「〜ている」「〜てみる」「〜てほしい」程度の文型はOK
- 返答は1〜2文。わかりやすくシンプルに
- あまり難しい表現は避ける`
	default: // N5 초급
		prompt += `

【言語レベル: N5初級】
- 非常に短くシンプルな文で話してください（1文は15字以内が理想）
- 難しい漢字はひらがなで読み方を添える: 例「好（す）き」「今日（きょう）」「面白（おもしろ）い」
- 難しい文法は使わない（〜ている / 〜ておく / 〜させる 等は禁止）
- 基本語彙のみ使用: すごい、好き、面白い、楽しい、わかる 等
- 返答は1文のみ。とにかく短く、ゆっくり`
	}

	if len(todaySentences) > 0 {
		prompt += "\n\n【学習サポート】\n以下は今日ユーザーが学習中の文章です。あなた自身はこれらをそのまま使わず、ユーザーが自然にこれらを言いたくなるような話題や質問を投げかけてください：\n"
		for _, s := range todaySentences {
			prompt += fmt.Sprintf("  - %s\n", s)
		}
	}

	return prompt
}

// levelSuggestionHint 레벨별 제안 생성 힌트 문자열 반환
func levelSuggestionHint(userLevel int) string {
	switch userLevel {
	case 3:
		return "N3 중상급 (자연스러운 표현, 10~25자)"
	case 4:
		return "N4 중급 (기본 문형, 10~20자)"
	default:
		return "N5 초급 (매우 짧고 단순한 표현, 5~15자)"
	}
}
