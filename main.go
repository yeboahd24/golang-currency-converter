package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
)

// CurrencyResponse represents the response from the currency conversion API
type CurrencyResponse struct {
    Success bool `json:"success"`
    Terms   string `json:"terms"`
    Privacy string `json:"privacy"`
    Query   struct {
        From   string `json:"from"`
        To     string `json:"to"`
        Amount int    `json:"amount"`
    } `json:"query"`
    Info struct {
        Timestamp int64   `json:"timestamp"`
        Quote     float64 `json:"quote"`
    } `json:"info"`
    Result float64 `json:"result"`
}

func main() {
    // Load environment variables from .env file
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    r := gin.Default()
    gin.SetMode(gin.ReleaseMode)

    // r.StaticFS("/home", http.Dir("./template/index.html"))
    r.StaticFile("/", "./template/index.html")

    r.GET("/convert", convertHandler)
    r.GET("/currencies", getCurrenciesHandler)
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080" // Default port if not specified
    }

    r.Run(":" + port)
}

func convertHandler(c *gin.Context) {
    fromCurrency := c.Query("from")
    toCurrency := c.Query("to")
    amount := c.Query("amount")

    if fromCurrency == "" || toCurrency == "" || amount == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required query parameters"})
        return
    }

    apiKey := os.Getenv("CURRENCY_API_KEY")
    if apiKey == "" {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "API key is not configured"})
        return
    }

    convertedCurrency, err := getConvertedCurrency(apiKey, fromCurrency, toCurrency, amount)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, convertedCurrency)
}

func getConvertedCurrency(apiKey, from, to, amount string) (*CurrencyResponse, error) {
    url := fmt.Sprintf("http://api.currencylayer.com/convert?access_key=%s&from=%s&to=%s&amount=%s", apiKey, from, to, amount)

    resp, err := http.Get(url)
    if err != nil {
        return nil, errors.New("failed to request currency conversion")
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API returned non-OK status: %s", resp.Status)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, errors.New("failed to read response body")
    }

    var currencyResponse CurrencyResponse
    if err := json.Unmarshal(body, &currencyResponse); err != nil {
        return nil, errors.New("failed to unmarshal response body")
    }

    return &currencyResponse, nil
}


// getCurrenciesHandler fetches currency list and returns it
func getCurrenciesHandler(c *gin.Context) {
    apiKey := os.Getenv("CURRENCY_API_KEY")
    if apiKey == "" {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "API key is not configured"})
        return
    }

    // Fetch the currency list from CurrencyLayer API
    resp, err := http.Get(fmt.Sprintf("http://api.currencylayer.com/list?access_key=%s", apiKey))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch currency list"})
        return
    }
    defer resp.Body.Close()

    // Read and forward the API response
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
        return
    }
    var result interface{}
    json.Unmarshal(body, &result)
    c.JSON(http.StatusOK, result)
}

