package router

import (
	"database/sql"
	controller "myapp/API"

	"github.com/gorilla/mux"
)

// InitRoutes initializes and returns the router with all the routes defined
func InitRoutes(db *sql.DB) *mux.Router {
	router := mux.NewRouter()

	// Routes for authentication
	router.HandleFunc("/register", controller.RegisterUser).Methods("POST")
	router.HandleFunc("/login", controller.LoginUser).Methods("POST")
	router.HandleFunc("/", controller.Test).Methods("GET")
	router.HandleFunc("/random", controller.GenerateLotteryHandler(db)).Methods("GET")
	router.HandleFunc("/win_lotto", controller.DrawPrizes).Methods("POST")
	router.HandleFunc("/win_lotto_all", controller.DrawPrizesAll).Methods("POST")
	router.HandleFunc("/buy_lottery", controller.BuyLottery).Methods("POST")
	router.HandleFunc("/get_lotto_for_buy", controller.GetUnpurchasedLotteriesHandler(db)).Methods("GET")
	router.HandleFunc("/check_lotto", controller.CheckUserLotteryResultsHandler(db)).Methods("POST")
	router.HandleFunc("/reward", controller.RewardPrize(db)).Methods("POST")
	router.HandleFunc("/reset", controller.ResetHandler(db)).Methods("POST")
	router.HandleFunc("/get1", controller.GetaUser).Methods("POST")
	router.HandleFunc("/lotto_buy_finish", controller.GetPurchasedLotteriesByUID).Methods("POST")
	router.HandleFunc("/lotto_buy_finish2", controller.GetPurchasedLotteriesByUID2).Methods("POST")
	router.HandleFunc("/get_all_lotto", controller.GetAllLotteriesHandler(db)).Methods("GET")
	router.HandleFunc("/get_win_prize", controller.GetAllWinningNumbers).Methods("GET")
	return router
}
