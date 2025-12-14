package data

import (
	"fmt"
	"log"
	"math"
)

type TextDataSearchCriteria struct {
	Category int
	MinIndex int
	MaxIndex int
	Between  bool
}

func (sc TextDataSearchCriteria) Query() string {
	if sc.MinIndex > sc.MaxIndex {
		log.Fatalf("minimum index cannot be higher than max index")
	}

	return buildQuery("t.'index', text", sc)
}

func (sc TextDataSearchCriteria) CountQuery() string {
	if sc.MinIndex > sc.MaxIndex {
		log.Fatalf("minimum index cannot be higher than max index")
	}

	return buildQuery("COUNT(*)", sc)
}

func buildQuery(selectStatement string, sc TextDataSearchCriteria) string {
	query := fmt.Sprintf("SELECT %s FROM text_data as t WHERE (t.category = %d)", selectStatement, sc.Category)
	queryBetween := fmt.Sprintf("AND (t.'index' BETWEEN %d AND %d)", sc.MinIndex, sc.MaxIndex)
	queryNotBetween := fmt.Sprintf("AND (t.'index' NOT BETWEEN %d AND %d)", sc.MinIndex, sc.MaxIndex)

	if sc.MinIndex == 0 && sc.MaxIndex == 0 {
		return query
	} else if sc.Between {
		return fmt.Sprintf("%s %s;", query, queryBetween)
	} else {
		return fmt.Sprintf("%s %s;", query, queryNotBetween)
	}
}

// text_data category IDs
var (
	// common                      = nil
	// misc                        = nil
	umaTitle = [...]TextDataSearchCriteria{
		{5, 0, 0, false},
	}
	charName = [...]TextDataSearchCriteria{
		{6, 0, 0, false},   // Ingame (during a career run), Director Akikawa, Trainer Kiryuin
		{77, 0, 0, false},  // Support cards
		{78, 0, 0, false},  // Support cards (for Logs?)
		{170, 0, 0, false}, // 6, but replaced to Yayoi Akikawa + Aoi Kiryuin
		{182, 0, 0, false}, // 6, but replaced to Yayoi Akikawa + Aoi Kiryuin (for Logs?)
	}
	itemName = [...]TextDataSearchCriteria{
		{23, 0, 0, false},
	}
	itemDesc = [...]TextDataSearchCriteria{
		{24, 0, 0, false},
	}
	itemGroups = [...]TextDataSearchCriteria{
		{133, 0, 0, false},
	}
	skillName = [...]TextDataSearchCriteria{
		{47, 0, 0, false},             // Skills
		{147, 2000100, 3000100, true}, // Skill sparks
	}
	skillDesc = [...]TextDataSearchCriteria{
		{48, 0, 0, false},
	}
	raceName = [...]TextDataSearchCriteria{
		{28, 20000, 30000, false}, // Full
		{29, 0, 0, false},         // Abbreviated
		{32, 0, 0, false},
		{33, 0, 0, false},
		{36, 0, 0, false},
		{111, 0, 0, false},
		{218, 0, 0, false},
	}
	raceTracks = [...]TextDataSearchCriteria{ // Hanshin, Sapporo, Nakayama, etc
		{31, 0, 0, false}, // Full
		{35, 0, 0, false}, // Abbreviated
	}
	eventTitles = [...]TextDataSearchCriteria{
		{189, 0, 0, false}, // Full
		{214, 0, 0, false}, // Abbreviated
	}
	storyTitlesMainstory = [...]TextDataSearchCriteria{
		{94, 0, 0, false},
	}
	storyTitlesAnnivstory = [...]TextDataSearchCriteria{
		{222, 0, 0, false},
	}
	storyTitlesEvents = [...]TextDataSearchCriteria{
		{191, 0, 0, false},
	}
	storyTitlesUmas = [...]TextDataSearchCriteria{
		{92, 0, 0, false},
	}
	scenarioNames = [...]TextDataSearchCriteria{
		{119, 0, 0, false}, {237, 0, 0, false},
	}
	trainerTitle = [...]TextDataSearchCriteria{
		{65, 0, 0, false}, // Profile title (CM epithets and other shit basically)
	}
	trainerTitleRequirements = [...]TextDataSearchCriteria{
		{66, 0, 0, false},
	}
	missions = [...]TextDataSearchCriteria{
		{67, 0, 0, false}, {190, 0, 0, false},
	}
	supportTitle = [...]TextDataSearchCriteria{
		{76, 0, 0, false},
	}
	advice = [...]TextDataSearchCriteria{
		{97, 0, 0, false},
	}
	umaEpithet = [...]TextDataSearchCriteria{
		{130, 0, 0, false},
	}
	umaEpithetColored = [...]TextDataSearchCriteria{
		{247, 0, 0, false},
	}
	umaEpithetRequirements = [...]TextDataSearchCriteria{
		{131, 0, 0, false},
	}
	pvpRawScoreName = [...]TextDataSearchCriteria{ // TT score calc text
		{140, 0, 0, false},
	}
	pvpRawScoreDesc = [...]TextDataSearchCriteria{
		{141, 0, 0, false},
	}
	conditionsName = [...]TextDataSearchCriteria{ // Status effects during career
		{142, 0, 0, false},
	}
	conditionsDesc = [...]TextDataSearchCriteria{
		{143, 0, 0, false},
	}
	pvpScoreBonusName = [...]TextDataSearchCriteria{ // TT ace, opponent rating, streaks, support bonus text
		{148, 0, 0, false},
	}
	storyEventMissions = [...]TextDataSearchCriteria{
		{190, 0, 0, false},
	}
	supportEffectUniqueDesc = [...]TextDataSearchCriteria{ // Support card unique effect
		{155, 0, 0, false},
	}
	itemAcquisitionMethodsShop = [...]TextDataSearchCriteria{
		{25, 0, 0, false},
	}
	factorDesc = [...]TextDataSearchCriteria{
		{172, 0, 0, false}, // Spark descriptions
	}
	supportBonus = [...]TextDataSearchCriteria{
		{186, 0, 0, false},
	}
	help = [...]TextDataSearchCriteria{
		{63, 0, 0, false},
	}
	predictions = [...]TextDataSearchCriteria{
		{27, 0, 0, false},
	}
	campaigns = [...]TextDataSearchCriteria{
		{187, 0, 0, false},
	}
	campaignsDesc = [...]TextDataSearchCriteria{
		{188, 0, 0, false},
	}
	gachaInfo = [...]TextDataSearchCriteria{
		{13, 0, 0, false},
	}
	gachaNames = [...]TextDataSearchCriteria{
		{26, 0, 0, false},
		{327, 0, 0, false},
	}
	presentsDesc = [...]TextDataSearchCriteria{
		{64, 0, 0, false},
	}
	itemUmaPieces = [...]TextDataSearchCriteria{
		{113, 0, 0, false},
	}
	umaProfileIntro = [...]TextDataSearchCriteria{
		{163, 0, 0, false},
	}
	umaProfileTagline = [...]TextDataSearchCriteria{
		{144, 0, 0, false},
	}
	umaProfileWeight = [...]TextDataSearchCriteria{
		{9, 0, 0, false},
	}
	umaProfileShoesize = [...]TextDataSearchCriteria{
		{168, 0, 0, false},
	}
	umaProfileStrengths = [...]TextDataSearchCriteria{
		{164, 0, 0, false},
	}
	umaProfileWeaknesses = [...]TextDataSearchCriteria{
		{165, 0, 0, false},
	}
	umaProfileEars = [...]TextDataSearchCriteria{
		{166, 0, 0, false},
	}
	umaProfileTail = [...]TextDataSearchCriteria{
		{167, 0, 0, false},
	}
	umaProfileFamily = [...]TextDataSearchCriteria{
		{169, 0, 0, false},
	}
	loginBonus = [...]TextDataSearchCriteria{
		{70, 0, 0, false},
	}
	specialTransferName = [...]TextDataSearchCriteria{
		{215, 0, 0, false},
	}
	specialTransferDesc = [...]TextDataSearchCriteria{
		{216, 0, 0, false},
	}
	specialTransferRequirements = [...]TextDataSearchCriteria{
		{217, 0, 0, false},
	}
	specialTransferThanks = [...]TextDataSearchCriteria{
		{220, 0, 0, false},
	}
	fourkomaTitles = [...]TextDataSearchCriteria{
		{69, 0, 7999, true},
	}
	tazunaTips = [...]TextDataSearchCriteria{
		{69, 8000, 8999, true},
	}
	umaSecrets = [...]TextDataSearchCriteria{
		{69, 9000, math.MaxInt32, false},
	}
	supportEffectUniqueName = [...]TextDataSearchCriteria{
		{150, 0, 0, false},
	}
	supportEffectName = [...]TextDataSearchCriteria{
		{151, 0, 0, false},
		{294, 0, 0, false},
	}
	supportEffectDesc = [...]TextDataSearchCriteria{
		{154, 0, 0, false},
	}
	supportFullName = [...]TextDataSearchCriteria{ // Support card full name
		{75, 0, 0, false},
	}
	tutorialText = [...]TextDataSearchCriteria{
		{3, 0, 0, false},
	}
	songDesc = [...]TextDataSearchCriteria{
		{128, 0, 0, false},
	}
	staticUmaProfile = [...]TextDataSearchCriteria{
		{175, 0, 0, false},
	}
	jukeboxComments = [...]TextDataSearchCriteria{
		{228, 0, 0, false},
	}
	scenarioDesc = [...]TextDataSearchCriteria{
		{120, 0, 0, false},
	}
	scenarioItemName = [...]TextDataSearchCriteria{
		{225, 0, 0, false},
	}
	scenarioItemDescLong = [...]TextDataSearchCriteria{
		{226, 0, 0, false},
	}
	scenarioItemDescShort = [...]TextDataSearchCriteria{
		{238, 0, 0, false},
	}
	mobName = [...]TextDataSearchCriteria{
		{59, 0, 0, false},
	}
	scenarioEffects = [...]TextDataSearchCriteria{
		{198, 0, 0, false},
		{199, 0, 0, false},
		{200, 0, 0, false},
	}
	lessonNames = [...]TextDataSearchCriteria{
		{209, 0, 0, false},
	}
	lessonTalentBonus = [...]TextDataSearchCriteria{
		{207, 0, 0, false},
	}
	lessonLiveBonus = [...]TextDataSearchCriteria{
		{208, 0, 0, false},
	}
	seiyuu = [...]TextDataSearchCriteria{
		{7, 0, 0, false},
	}
	cardStories = [...]TextDataSearchCriteria{
		{88, 0, 0, false},
	}
	meekChallenges = [...]TextDataSearchCriteria{
		{267, 0, 0, false},
	}
	strollLocations = [...]TextDataSearchCriteria{
		{268, 0, 0, false},
	}
	valentineGiftNames = [...]TextDataSearchCriteria{
		{273, 0, 0, false},
	}
	valentineGiftNotes = [...]TextDataSearchCriteria{
		{271, 0, 0, false},
		{272, 0, 0, false},
	}
	knowledgeEffects = [...]TextDataSearchCriteria{
		{296, 0, 0, false},
	}
	knowledgeBuffs = [...]TextDataSearchCriteria{
		{298, 0, 0, false},
	}
	skillEvolveCond = [...]TextDataSearchCriteria{
		{290, 0, 0, false},
	}
	exchangeNames = [...]TextDataSearchCriteria{
		{39, 0, 0, false},
	}
	cashshopNames = [...]TextDataSearchCriteria{
		{49, 0, 0, false},
	}
	cashshopDesc = [...]TextDataSearchCriteria{
		{96, 0, 100, true},
	}
	outfits = [...]TextDataSearchCriteria{
		{14, 0, 0, false},
	}
	liveNames = [...]TextDataSearchCriteria{
		{16, 0, 0, false},
	}
	loadscreenTitles = [...]TextDataSearchCriteria{
		{68, 0, 0, false},
	}
	umaRank = [...]TextDataSearchCriteria{ // Fan ranking in career (debut, maiden, ..., legend)
		{121, 0, 0, false},
	}
	trainingComments = [...]TextDataSearchCriteria{ // Etsuko comments during interviews lole
		{139, 0, 0, false},
	}
	errDesc = [...]TextDataSearchCriteria{
		{1, 0, 0, false},
	}
	errName = [...]TextDataSearchCriteria{
		{2, 0, 0, false},
	}
	trainingNames = [...]TextDataSearchCriteria{
		{55, 0, 0, false},
		{138, 0, 0, false},
	}
	hallPosters = [...]TextDataSearchCriteria{
		{196, 0, 0, false},
	}
	evSakuraPowerAreas = [...]TextDataSearchCriteria{
		{304, 0, 0, false}, {309, 0, 0, false},
	}
	lohMessages = [...]TextDataSearchCriteria{
		{297, 0, 0, false},
	}
	effectFields = [...]TextDataSearchCriteria{
		{293, 0, 0, false},
	}
	miscEvents = [...]TextDataSearchCriteria{
		{274, 0, 0, false},
		{283, 0, 0, false},
		{282, 0, 0, false},
		{281, 0, 0, false},
		{280, 0, 0, false},
		{279, 0, 0, false},
		{284, 0, 0, false},
		{285, 0, 0, false},
		{286, 0, 0, false},
		{287, 0, 0, false},
		{288, 0, 0, false},
		{289, 0, 0, false},
		{253, 0, 0, false},
		{251, 0, 0, false},
		{248, 0, 0, false},
	}
	specialGacha = [...]TextDataSearchCriteria{
		{277, 0, 0, false},
	}
	lohRequirements = [...]TextDataSearchCriteria{
		{275, 0, 0, false},
	}
	doubleRequirements = [...]TextDataSearchCriteria{
		{257, 0, 0, false},
	}
	grandLive = [...]TextDataSearchCriteria{
		{250, 0, 0, false},
	}
	specialMissions = [...]TextDataSearchCriteria{
		{252, 0, 0, false},
	}
	fujiShowtime = [...]TextDataSearchCriteria{
		{246, 0, 0, false},
		{244, 0, 0, false},
		{245, 0, 0, false},
		{239, 0, 0, false},
		{236, 0, 0, false},
		{235, 0, 0, false},
		{234, 0, 0, false},
	}
	scenarioMessages = [...]TextDataSearchCriteria{
		{240, 0, 0, false},
	}
	cmMagazineQuestions = [...]TextDataSearchCriteria{
		{231, 0, 0, false},
	}
	cmMagazineComments = [...]TextDataSearchCriteria{
		{232, 0, 0, false},
	}
	cmMagazineHeaders = [...]TextDataSearchCriteria{
		{230, 0, 0, false},
	}
	foreignApts = [...]TextDataSearchCriteria{
		{310, 0, 0, false},
	}
	foreignAptEffects = [...]TextDataSearchCriteria{
		{317, 0, 0, false},
		{318, 0, 0, false},
	}
	foreignAptEffectsColored = [...]TextDataSearchCriteria{
		{311, 0, 0, false},
		{314, 0, 0, false},
	}
	foreignAptReqs = [...]TextDataSearchCriteria{
		{313, 0, 0, false},
	}
	larcChallenges = [...]TextDataSearchCriteria{
		{315, 0, 0, false},
	}
	larcChallengeDebuffs = [...]TextDataSearchCriteria{
		{316, 0, 0, false},
	}
	larcEvoAlts = [...]TextDataSearchCriteria{
		{326, 0, 0, false},
	}
	ssmatchEffects = [...]TextDataSearchCriteria{
		{322, 0, 0, false},
	}
	uafHeatupEffects = [...]TextDataSearchCriteria{
		{330, 0, 0, false},
	}
	uafEpithets = [...]TextDataSearchCriteria{
		{345, 0, 0, false},
	}
	gffDishes = [...]TextDataSearchCriteria{
		{332, 0, 0, false},
		{349, 0, 0, false},
		{359, 0, 0, false},
	}
	gffDishesDesc = [...]TextDataSearchCriteria{
		{353, 0, 0, false},
	}
	gffEffects = [...]TextDataSearchCriteria{
		{333, 0, 0, false},
		{358, 0, 0, false},
	}
	gffMessages = [...]TextDataSearchCriteria{
		{354, 0, 0, false},
		{352, 0, 0, false},
	}
	gffProduce = [...]TextDataSearchCriteria{
		{331, 0, 0, false},
	}
	gffProduceLv = [...]TextDataSearchCriteria{
		{334, 0, 0, false},
	}
)
