package questions

// func (s *StrategyLink) answerSecurityQuestions(w http.ResponseWriter, r *http.Request, req *Request) {
// 	for _, question := range req.RecoveredIdentity.RecoverySecurityAnswers {
// 		answer := r.PostForm.Get(securityQuestionPrefix + "." + question.Key)
// 		if len(answer) == 0 {
// 			s.handleError(w, r, req, schema.NewRequiredError("#/"+securityQuestionPrefix, question.Key))
// 			return
// 		}
//
// 		s.d.RecoveryManager().CompareSecurityQuestions(r.Context(),question,answer)
// 	}
// }
