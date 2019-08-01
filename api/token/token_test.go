package token_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/base64"
	"errors"
	. "github.com/globocom/huskyCI/api/token"
	"github.com/globocom/huskyCI/api/types"
	"time"
)

type FakeExternal struct {
	expectedURL              string
	expectedValidateError    error
	expectedToken            string
	expectedGenerateError    error
	expectedTime             time.Time
	expectedStoreAccessError error
	expectedAccessToken      types.AccessToken
	expectedFindAccessError  error
	expectedFindRepoError    error
}

func (fE *FakeExternal) ValidateURL(url string) (string, error) {
	return fE.expectedURL, fE.expectedValidateError
}

func (fE *FakeExternal) GenerateToken() (string, error) {
	return fE.expectedToken, fE.expectedGenerateError
}

func (fE *FakeExternal) GetTimeNow() time.Time {
	return fE.expectedTime
}

func (fE *FakeExternal) StoreAccessToken(accessToken types.AccessToken) error {
	return fE.expectedStoreAccessError
}

func (fE *FakeExternal) FindAccessToken(token, repositoryURL string) (types.AccessToken, error) {
	return fE.expectedAccessToken, fE.expectedFindAccessError
}

func (fE *FakeExternal) FindRepoURL(repositoryURL string) error {
	return fE.expectedFindRepoError
}

var _ = Describe("Token", func() {
	Context("When URL validation returns an error", func() {
		It("Should return the same error and an empty struct", func() {
			fakeExt := FakeExternal{
				expectedURL:           "",
				expectedValidateError: errors.New("URL is not valid"),
			}
			tokenGen := TokenHandler{
				External: &fakeExt,
			}
			accessToken, err := tokenGen.GenerateAccessToken(types.TokenRequest{
				RepositoryURL: "myRepo.com",
			})
			Expect(accessToken).To(Equal(types.AccessToken{}))
			Expect(err).To(Equal(errors.New("URL is not valid")))
		})
	})
	Context("When GenerateToken returns an error", func() {
		It("Should return the same error and an empty struct", func() {
			fakeExt := FakeExternal{
				expectedURL:           "",
				expectedValidateError: nil,
				expectedToken:         "",
				expectedGenerateError: errors.New("Failed to generate token"),
			}
			tokenGen := TokenHandler{
				External: &fakeExt,
			}
			accessToken, err := tokenGen.GenerateAccessToken(types.TokenRequest{
				RepositoryURL: "myRepo.com",
			})
			Expect(accessToken).To(Equal(types.AccessToken{}))
			Expect(err).To(Equal(errors.New("Failed to generate token")))
		})
	})
	Context("When StoreAccessToken return an error", func() {
		It("Should return the same error and an empty struct", func() {
			fakeExt := FakeExternal{
				expectedURL:              "https://www.github.com/myProject",
				expectedValidateError:    nil,
				expectedToken:            base64.URLEncoding.EncodeToString([]byte("RandomValue")),
				expectedGenerateError:    nil,
				expectedTime:             time.Now(),
				expectedStoreAccessError: errors.New("Failed to store access token in DB"),
			}
			tokenGen := TokenHandler{
				External: &fakeExt,
			}
			accessToken, err := tokenGen.GenerateAccessToken(types.TokenRequest{
				RepositoryURL: "github.com/myProject",
			})
			Expect(accessToken).To(Equal(types.AccessToken{}))
			Expect(err).To(Equal(errors.New("Failed to store access token in DB")))
		})
	})
	Context("When a valid repo URL and a new token are generated", func() {
		It("Should return the expected accessToken struct and a nil error", func() {
			fakeExt := FakeExternal{
				expectedURL:              "https://www.github.com/myProject",
				expectedValidateError:    nil,
				expectedToken:            base64.URLEncoding.EncodeToString([]byte("RandomValue")),
				expectedGenerateError:    nil,
				expectedTime:             time.Now(),
				expectedStoreAccessError: nil,
			}
			tokenGen := TokenHandler{
				External: &fakeExt,
			}
			accessToken, err := tokenGen.GenerateAccessToken(types.TokenRequest{
				RepositoryURL: "github.com/myProject",
			})
			Expect(accessToken).To(Equal(types.AccessToken{
				HuskyToken: fakeExt.expectedToken,
				URL:        fakeExt.expectedURL,
				IsValid:    true,
				CreatedAt:  fakeExt.expectedTime,
			}))
			Expect(err).To(BeNil())
		})
	})
	Context("When ValidateURL returns an error", func() {
		It("Should return the same error", func() {
			fakeExt := FakeExternal{
				expectedURL:           "",
				expectedValidateError: errors.New("Error during URL validation"),
			}
			tokenVal := TokenHandler{
				External: &fakeExt,
			}
			Expect(tokenVal.ValidateToken("MyToken", "myProject")).To(Equal(fakeExt.expectedValidateError))
		})
	})
	Context("When FindAccessToken returns an error", func() {
		It("Should return the same error as expected", func() {
			fakeExt := FakeExternal{
				expectedAccessToken:     types.AccessToken{},
				expectedFindAccessError: errors.New("Could not find current access token"),
			}
			tokenVal := TokenHandler{
				External: &fakeExt,
			}
			Expect(tokenVal.ValidateToken("MyToken", "myProject")).To(Equal(fakeExt.expectedFindAccessError))
		})
	})
	Context("When FindAccessToken returns a invalid access token", func() {
		It("Should return the expected error", func() {
			fakeExt := FakeExternal{
				expectedAccessToken: types.AccessToken{
					IsValid: false,
				},
				expectedFindAccessError: nil,
			}
			tokenVal := TokenHandler{
				External: &fakeExt,
			}
			Expect(tokenVal.ValidateToken("MyToken", "myProject")).To(Equal(errors.New("Access token is invalid")))
		})
	})
	Context("When FindAccessToken returns a valid access token", func() {
		It("Should return a nil error", func() {
			fakeExt := FakeExternal{
				expectedAccessToken: types.AccessToken{
					HuskyToken: "MyToken",
					IsValid:    true,
					URL:        "myProject",
					CreatedAt:  time.Now(),
				},
			}
			tokenVal := TokenHandler{
				External: &fakeExt,
			}
			Expect(tokenVal.ValidateToken("MyToken", "myProject")).To(BeNil())
		})
	})
	Describe("VerifyRepo", func() {
		Context("When ValidateURL returns an error", func() {
			It("Should return the same error", func() {
				fakeExt := FakeExternal{
					expectedURL:           "",
					expectedValidateError: errors.New("Repository does not have a valid format"),
				}
				verRepo := TokenHandler{
					External: &fakeExt,
				}
				Expect(verRepo.VerifyRepo("MyRepo")).To(Equal(fakeExt.expectedValidateError))
			})
		})
		Context("When FindRepoURL returns something", func() {
			It("Should return the same error if it has returned an error", func() {
				fakeExt := FakeExternal{
					expectedURL:           "https://www.github.com/myProject",
					expectedValidateError: nil,
					expectedFindRepoError: errors.New("Repository URL not found"),
				}
				verRepo := TokenHandler{
					External: &fakeExt,
				}
				Expect(verRepo.VerifyRepo("MyRepo")).To(Equal(fakeExt.expectedFindRepoError))
			})
			It("Should return nil if the a repository URL was found", func() {
				fakeExt := FakeExternal{
					expectedURL:           "https://www.github.com/myProject",
					expectedValidateError: nil,
					expectedFindRepoError: nil,
				}
				verRepo := TokenHandler{
					External: &fakeExt,
				}
				Expect(verRepo.VerifyRepo("MyRepo")).To(BeNil())
			})
		})
	})
})
