package response

// OnboardingLanguage is the compact language representation used by onboarding.
type OnboardingLanguage struct {
	ID   int    `json:"id" example:"1"`
	Code string `json:"code" example:"en"`
	Name string `json:"name" example:"English"`
	Flag string `json:"flag" example:"🇬🇧"`
} // @name response.OnboardingLanguage

// OnboardingTopic is an active topic localized for the selected translation language.
type OnboardingTopic struct {
	ID   int    `json:"id" example:"1"`
	Slug string `json:"slug" example:"travel"`
	Icon string `json:"icon" example:"https://example.com/travel.svg"`
	Name string `json:"name" example:"Travel"`
} // @name response.OnboardingTopic

// Onboarding contains the catalogs needed to render the onboarding screen.
type Onboarding struct {
	LearningLanguages    []OnboardingLanguage `json:"learningLanguages"`
	TranslationLanguages []OnboardingLanguage `json:"translationLanguages"`
	Topics               []OnboardingTopic    `json:"topics"`
} // @name response.Onboarding
