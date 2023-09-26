package mongohelpers
import (
	"context"
	"fmt"
	"strings"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"indicatorsAPP/config"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
)
/*
					########################### 
				ALL Mongo Operations are defined here
*/
// Bulk Write
// Example Usage :
/*
var bulkUpdateModels []mongo.WriteModel
// Create the update model with the filter and update documents
updateModel := mongo.NewUpdateOneModel()
updateModel.SetFilter(bson.M{})
updateModel.SetUpdate(bson.M{$set:{bson.M{}}})
updateModel.SetUpsert(true) // Set upsert to true to insert if not found
// Append the update model to the bulkUpdateModels slice
bulkUpdateModels = append(bulkUpdateModels, updateModel)
*/
func MongoBulkWrite(collectionName string, bulkUpdateModels []mongo.WriteModel) (*mongo.BulkWriteResult,error) {
	collection := config.GetMongoDB().Collection(collectionName)
	// Create the options for the bulk write operation (e.g., ordered = false for unordered bulk write)
	bulkWriteOptions := options.BulkWrite().SetOrdered(false)
	// Perform the bulk write operation on the MongoDB collection
	bulkWriteResult, err := collection.BulkWrite(context.Background(), bulkUpdateModels, bulkWriteOptions)
	if err != nil {
		return nil , fmt.Errorf("failed to perform bulk write in MongoDB: %w", err)
	}
	return bulkWriteResult, nil
}

// Find and Update single Document use Upsert accordingly
func MongoUpdateOne(collectionName string, filters bson.M, update bson.M,upsert bool) error {
	collection := config.GetMongoDB().Collection(collectionName)
	// Define options for the update operation.
	updateOptions := options.Update().SetUpsert(upsert)
	// Perform the update operation.
	_, err := collection.UpdateOne(context.Background(), filters, update,updateOptions)
	if err != nil {
		return  err
	}

	return  nil
}

// Find and Update Multiple Documents
func MongoUpdateMany(collectionName string, filters bson.M, update bson.M,upsert bool) error {
	collection := config.GetMongoDB().Collection(collectionName)
	// Define options for the update operation.
	updateOptions := options.Update().SetUpsert(upsert)
	// Perform the update operation.
	var err error
	if len(filters) == 0 {
		// If filters are empty, update all documents
		_, err = collection.UpdateMany(context.Background(), bson.M{}, update, updateOptions)
	} else {
		_, err = collection.UpdateMany(context.Background(), filters, update, updateOptions)
	}
	if err != nil {
		return  err
	}

	return  nil
}

func MongoFindOne(collectionName string, filters bson.M,projection bson.M) (bson.M,error){
	collection := config.GetMongoDB().Collection(collectionName)
	// mongo options
	findOptions := options.FindOne()
	if projection != nil {
		findOptions.SetProjection(projection)
	}
	// Perform the find operation.
	var doc bson.M
	if err := collection.FindOne(context.Background(), filters,findOptions).Decode(&doc); err != nil {
		return bson.M{}, err
	}
	return doc, nil
}
func MongoFind(collectionName string, filters bson.M,projection bson.M, limit int64, sortOrder int, sortBy string) ([]bson.M, error) {
	collection := config.GetMongoDB().Collection(collectionName)
	// Define options for the find operation.
	findOptions := options.Find()
	// Handle optional parameters
	// Projection for fields
	if projection != nil {
		findOptions.SetProjection(projection)
	}
	// limit 
	if limit > 0 {
		findOptions.SetLimit(limit)
	}
	// sortBy
	if sortBy != "" {
		sortOrderValue := 1 // 1 for ascending, -1 for descending.
		if sortOrder == -1 {
			sortOrderValue = -1
		}
		findOptions.SetSort(map[string]int{sortBy: sortOrderValue})
	}
	// Perform the find operation.
	cursor, err := collection.Find(context.Background(), filters, findOptions)
	if err != nil {
		fmt.Println("err",err)
		return nil, err
	}
	defer cursor.Close(context.Background())
	// Process the results.
	var documents []bson.M
	for cursor.Next(context.Background()) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}
	return documents, nil
}

// Delete One
func MongoDeleteOne( collectionName string, filters bson.M) error {
	collection := config.GetMongoDB().Collection(collectionName)
	// Perform the delete operation.
	_, err := collection.DeleteOne(context.Background(), filters)
	return err
}
// DeleteMany
func MongoDeleteMany( collectionName string, filters bson.M) error {
	collection := config.GetMongoDB().Collection(collectionName)
	// Perform the delete operation.
	_, err := collection.DeleteMany(context.Background(), filters)
	return err
}
// Insert One
func MongoInsertOne(collectionName string, data bson.M) (primitive.ObjectID,error) {
	collection := config.GetMongoDB().Collection(collectionName)
	
	result , err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	insertDocID := (result.InsertedID).(primitive.ObjectID)
	return insertDocID , nil
}


// Aggregate
func  MongoAggregate( collectionName string, pipeline []bson.M) ([]bson.M , error) {
	var results []bson.M
	collection := config.GetMongoDB().Collection(collectionName)
	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

// Count Mongo Documents
func MongoCountDocuments( collectionName string, filters bson.M) (int64, error) {
	collection := config.GetMongoDB().Collection(collectionName)
	countOptions := options.Count()
	count, err := collection.CountDocuments(context.Background(), filters, countOptions)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ConvertStringToObjectId converts a string representation of an ObjectID to primitive.ObjectID.
func ConvertStringToObjectId(id string) (primitive.ObjectID, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return objID, nil
}


// Convert To Object ID
func ConvertToMongoID(input interface{}) (primitive.ObjectID, error) {
	switch t := input.(type) {
	case primitive.ObjectID:
		// If it's already an ObjectID, return it as is
		return input.(primitive.ObjectID), nil
	case string:
		// If it's a string, try converting it to an ObjectID
		if primitive.IsValidObjectID(t) {
			return primitive.ObjectIDFromHex(t)
		}
	}

	// If the input is neither an ObjectID nor a valid ObjectID string, return an error
	return primitive.ObjectID{}, fmt.Errorf("not a valid MongoDB ObjectID")
}

func ConvertObjectIDToString(doc map[string]interface{}, key string) (string, error) {
	// Check if the specified key exists in the map
	value, found := doc[key]
	if !found {
		return "", fmt.Errorf("key '%s' not found in the map", key)
	}
	// Type assertion to check if the value is already a primitive.ObjectID
	if objectID, ok := value.(primitive.ObjectID); ok {
		return objectID.Hex(), nil
	}

	// Value is not a primitive.ObjectID, let's check if it's a string representation
	if objectIDStr, ok := value.(string); ok {
		// Trim the "ObjectID()" prefix and suffix from the string representation
		objectIDStr = strings.TrimPrefix(objectIDStr, "ObjectID(")
		objectIDStr = strings.TrimSuffix(objectIDStr, ")")

		// Convert the string to a primitive.ObjectID
		objectID, err := primitive.ObjectIDFromHex(objectIDStr)
		if err != nil {
			return "", fmt.Errorf("key '%s' does not contain a valid ObjectID", key)
		}

		// Convert the ObjectID to a string
		return objectID.Hex(), nil
	}

	return "", fmt.Errorf("key '%s' does not contain a valid ObjectID", key)

}