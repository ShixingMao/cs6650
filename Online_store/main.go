package main

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

// Product struct matching the OpenAPI schema
type Product struct {
	ProductID    int    `json:"product_id"`
	SKU          string `json:"sku"`
	Manufacturer string `json:"manufacturer"`
	CategoryID   int    `json:"category_id"`
	Weight       int    `json:"weight"`
	SomeOtherID  int    `json:"some_other_id"`
}

// Thread-safe in-memory store
var (
	productStore = make(map[int]Product)
	storeLock    sync.RWMutex
)

func main() {
	router := gin.Default()

	// POST /products/:productId/details
	router.POST("/products/:productId/details", addProductDetails)

	// GET /products/:productId
	router.GET("/products/:productId", getProduct)

	router.Run(":8080")
}

// POST handler: Add or update product details
func addProductDetails(c *gin.Context) {
	productIdStr := c.Param("productId")
	productId, err := strconv.Atoi(productIdStr)
	if err != nil || productId < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_INPUT",
			"message": "Product ID must be a positive integer",
		})
		return
	}

	var product Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_INPUT",
			"message": "Invalid JSON body",
		})
		return
	}

	// Validate required fields
	if product.SKU == "" || product.Manufacturer == "" || product.CategoryID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_INPUT",
			"message": "Missing required fields",
		})
		return
	}

	// Store or update the product
	storeLock.Lock()
	productStore[productId] = product
	storeLock.Unlock()

	c.Status(http.StatusNoContent)
}

// GET handler: Retrieve product by ID
func getProduct(c *gin.Context) {
	productIdStr := c.Param("productId")
	productId, err := strconv.Atoi(productIdStr)
	if err != nil || productId < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_INPUT",
			"message": "Product ID must be a positive integer",
		})
		return
	}

	storeLock.RLock()
	product, exists := productStore[productId]
	storeLock.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "NOT_FOUND",
			"message": "Product not found",
		})
		return
	}

	c.JSON(http.StatusOK, product)
}
