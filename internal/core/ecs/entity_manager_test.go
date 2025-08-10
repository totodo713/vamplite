package ecs

import (
	"sync"
	"testing"
	"time"
)

// TestEntityManager_CreateEntity tests basic entity creation functionality.
func TestEntityManager_CreateEntity(t *testing.T) {
	em := NewDefaultEntityManager()

	t.Run("TC001: Create new entity", func(t *testing.T) {
		entity := em.CreateEntity()
		if entity == 0 {
			t.Error("Created entity ID should not be 0")
		}
	})

	t.Run("TC002: Created entity ID is valid", func(t *testing.T) {
		entity := em.CreateEntity()
		if entity <= 0 {
			t.Errorf("Created entity ID should be > 0, got %d", entity)
		}
	})

	t.Run("TC003: Sequential entities have unique IDs", func(t *testing.T) {
		entity1 := em.CreateEntity()
		entity2 := em.CreateEntity()
		if entity1 == entity2 {
			t.Errorf("Sequential entities should have unique IDs, got %d == %d", entity1, entity2)
		}
	})

	t.Run("TC004: Entity count increases correctly", func(t *testing.T) {
		initialCount := em.GetEntityCount()
		em.CreateEntity()
		newCount := em.GetEntityCount()
		if newCount != initialCount+1 {
			t.Errorf("Entity count should increase by 1, expected %d, got %d", initialCount+1, newCount)
		}
	})
}

// TestEntityManager_CreateEntityWithID tests entity creation with specific ID.
func TestEntityManager_CreateEntityWithID(t *testing.T) {
	em := NewDefaultEntityManager()

	t.Run("TC005: Create entity with specific ID", func(t *testing.T) {
		entityID := EntityID(100)
		err := em.CreateEntityWithID(entityID)
		if err != nil {
			t.Errorf("Should be able to create entity with specific ID, got error: %v", err)
		}
		if !em.IsValid(entityID) {
			t.Error("Created entity should be valid")
		}
	})

	t.Run("TC006: Error when creating entity with existing ID", func(t *testing.T) {
		entityID := EntityID(200)
		err1 := em.CreateEntityWithID(entityID)
		if err1 != nil {
			t.Errorf("First creation should succeed, got error: %v", err1)
		}

		err2 := em.CreateEntityWithID(entityID)
		if err2 == nil {
			t.Error("Second creation with same ID should return error")
		}
		if err2.(*ECSError).Code != ErrEntityAlreadyExists {
			t.Errorf("Expected ErrEntityAlreadyExists, got %v", err2)
		}
	})

	t.Run("TC007: Error when creating entity with invalid ID (0)", func(t *testing.T) {
		err := em.CreateEntityWithID(EntityID(0))
		if err == nil {
			t.Error("Creating entity with ID 0 should return error")
		}
		if err.(*ECSError).Code != ErrInvalidEntityID {
			t.Errorf("Expected ErrInvalidEntityID, got %v", err)
		}
	})
}

// TestEntityManager_DestroyEntity tests entity destruction functionality.
func TestEntityManager_DestroyEntity(t *testing.T) {
	em := NewDefaultEntityManager()

	t.Run("TC008: Destroy valid entity", func(t *testing.T) {
		entity := em.CreateEntity()
		err := em.DestroyEntity(entity)
		if err != nil {
			t.Errorf("Should be able to destroy valid entity, got error: %v", err)
		}
	})

	t.Run("TC009: Entity becomes invalid after destruction", func(t *testing.T) {
		entity := em.CreateEntity()
		_ = em.DestroyEntity(entity)
		if em.IsValid(entity) {
			t.Error("Entity should be invalid after destruction")
		}
	})

	t.Run("TC010: Error when destroying invalid entity", func(t *testing.T) {
		err := em.DestroyEntity(EntityID(99999))
		if err == nil {
			t.Error("Destroying invalid entity should return error")
		}
		if err.(*ECSError).Code != ErrEntityNotFound {
			t.Errorf("Expected ErrEntityNotFound, got %v", err)
		}
	})

	t.Run("TC011: Entity count decreases correctly", func(t *testing.T) {
		entity := em.CreateEntity()
		initialCount := em.GetEntityCount()
		_ = em.DestroyEntity(entity)
		newCount := em.GetEntityCount()
		if newCount != initialCount-1 {
			t.Errorf("Entity count should decrease by 1, expected %d, got %d", initialCount-1, newCount)
		}
	})
}

// TestEntityManager_IsValid tests entity validity checking.
func TestEntityManager_IsValid(t *testing.T) {
	em := NewDefaultEntityManager()

	t.Run("TC012: Newly created entity is valid", func(t *testing.T) {
		entity := em.CreateEntity()
		if !em.IsValid(entity) {
			t.Error("Newly created entity should be valid")
		}
	})

	t.Run("TC013: Destroyed entity is invalid", func(t *testing.T) {
		entity := em.CreateEntity()
		_ = em.DestroyEntity(entity)
		if em.IsValid(entity) {
			t.Error("Destroyed entity should be invalid")
		}
	})

	t.Run("TC014: Non-existent entity is invalid", func(t *testing.T) {
		if em.IsValid(EntityID(99999)) {
			t.Error("Non-existent entity should be invalid")
		}
	})

	t.Run("TC015: Entity ID 0 is invalid", func(t *testing.T) {
		if em.IsValid(EntityID(0)) {
			t.Error("Entity ID 0 should be invalid")
		}
	})
}

// TestEntityManager_RecycleEntity tests entity recycling functionality.
func TestEntityManager_RecycleEntity(t *testing.T) {
	em := NewDefaultEntityManager()

	t.Run("TC016: Add entity to recycle pool", func(t *testing.T) {
		entity := em.CreateEntity()
		_ = em.DestroyEntity(entity)
		err := em.RecycleEntity(entity)
		if err != nil {
			t.Errorf("Should be able to recycle destroyed entity, got error: %v", err)
		}
	})

	t.Run("TC017: Recycled ID is reused on new creation", func(t *testing.T) {
		entity1 := em.CreateEntity()
		em.DestroyEntity(entity1)
		em.RecycleEntity(entity1)

		entity2 := em.CreateEntity()
		if entity2 != entity1 {
			t.Errorf("Expected recycled ID %d to be reused, got %d", entity1, entity2)
		}
	})

	t.Run("TC018: Recycled count is tracked correctly", func(t *testing.T) {
		initialCount := em.GetRecycledCount()
		entity := em.CreateEntity()
		_ = em.DestroyEntity(entity)
		_ = em.RecycleEntity(entity)

		newCount := em.GetRecycledCount()
		if newCount != initialCount+1 {
			t.Errorf("Recycled count should increase by 1, expected %d, got %d", initialCount+1, newCount)
		}
	})

	t.Run("TC019: Error when recycling invalid entity", func(t *testing.T) {
		err := em.RecycleEntity(EntityID(99999))
		if err == nil {
			t.Error("Recycling invalid entity should return error")
		}
		if err.(*ECSError).Code != ErrEntityNotFound {
			t.Errorf("Expected ErrEntityNotFound, got %v", err)
		}
	})
}

// TestEntityManager_ClearRecycled tests recycled pool clearing.
func TestEntityManager_ClearRecycled(t *testing.T) {
	em := NewDefaultEntityManager()

	t.Run("TC020: Clear recycled pool", func(t *testing.T) {
		// Add some entities to recycle pool
		for i := 0; i < 5; i++ {
			entity := em.CreateEntity()
			_ = em.DestroyEntity(entity)
			_ = em.RecycleEntity(entity)
		}

		err := em.ClearRecycled()
		if err != nil {
			t.Errorf("Should be able to clear recycled pool, got error: %v", err)
		}
	})

	t.Run("TC021: Recycled count is zero after clear", func(t *testing.T) {
		entity := em.CreateEntity()
		_ = em.DestroyEntity(entity)
		_ = em.RecycleEntity(entity)

		_ = em.ClearRecycled()
		count := em.GetRecycledCount()
		if count != 0 {
			t.Errorf("Recycled count should be 0 after clear, got %d", count)
		}
	})

	t.Run("TC022: New entities get fresh IDs after clear", func(t *testing.T) {
		entity1 := em.CreateEntity()
		em.DestroyEntity(entity1)
		em.RecycleEntity(entity1)
		_ = em.ClearRecycled()

		entity2 := em.CreateEntity()
		if entity2 == entity1 {
			t.Errorf("After clearing recycled pool, new entity should get fresh ID, got reused ID %d", entity1)
		}
	})
}

// TestEntityManager_SetParent tests parent-child relationship setting.
func TestEntityManager_SetParent(t *testing.T) {
	em := NewDefaultEntityManager()

	t.Run("TC023: Set valid parent-child relationship", func(t *testing.T) {
		parent := em.CreateEntity()
		child := em.CreateEntity()

		err := em.SetParent(child, parent)
		if err != nil {
			t.Errorf("Should be able to set valid parent-child relationship, got error: %v", err)
		}
	})

	t.Run("TC024: Get parent returns correct parent", func(t *testing.T) {
		parent := em.CreateEntity()
		child := em.CreateEntity()
		_ = em.SetParent(child, parent)

		retrievedParent, exists := em.GetParent(child)
		if !exists {
			t.Error("Should be able to get parent of child entity")
		}
		if retrievedParent != parent {
			t.Errorf("Retrieved parent should be %d, got %d", parent, retrievedParent)
		}
	})

	t.Run("TC025: Get children returns correct children", func(t *testing.T) {
		parent := em.CreateEntity()
		child1 := em.CreateEntity()
		child2 := em.CreateEntity()

		_ = em.SetParent(child1, parent)
		_ = em.SetParent(child2, parent)

		children := em.GetChildren(parent)
		if len(children) != 2 {
			t.Errorf("Parent should have 2 children, got %d", len(children))
		}

		childMap := make(map[EntityID]bool)
		for _, child := range children {
			childMap[child] = true
		}

		if !childMap[child1] || !childMap[child2] {
			t.Error("Retrieved children should include both child1 and child2")
		}
	})

	t.Run("TC026: Error when setting circular reference", func(t *testing.T) {
		entity1 := em.CreateEntity()
		entity2 := em.CreateEntity()

		em.SetParent(entity2, entity1)
		err := em.SetParent(entity1, entity2)

		if err == nil {
			t.Error("Setting circular reference should return error")
		}
		if err.(*ECSError).Code != ErrCircularDependency {
			t.Errorf("Expected ErrCircularDependency, got %v", err)
		}
	})

	t.Run("TC027: Error when setting relationship with invalid entity", func(t *testing.T) {
		validEntity := em.CreateEntity()

		err1 := em.SetParent(EntityID(99999), validEntity)
		if err1 == nil {
			t.Error("Setting parent with invalid child should return error")
		}

		err2 := em.SetParent(validEntity, EntityID(99999))
		if err2 == nil {
			t.Error("Setting parent with invalid parent should return error")
		}
	})
}

// TestEntityManager_GetHierarchy tests hierarchy traversal functionality.
func TestEntityManager_GetHierarchy(t *testing.T) {
	em := NewDefaultEntityManager()

	t.Run("TC028: Get all descendants", func(t *testing.T) {
		// Create hierarchy: grandparent -> parent -> child
		grandparent := em.CreateEntity()
		parent := em.CreateEntity()
		child := em.CreateEntity()

		em.SetParent(parent, grandparent)
		_ = em.SetParent(child, parent)

		descendants := em.GetDescendants(grandparent)
		if len(descendants) != 2 {
			t.Errorf("Grandparent should have 2 descendants, got %d", len(descendants))
		}

		descendantMap := make(map[EntityID]bool)
		for _, desc := range descendants {
			descendantMap[desc] = true
		}

		if !descendantMap[parent] || !descendantMap[child] {
			t.Error("Descendants should include both parent and child")
		}
	})

	t.Run("TC029: Get all ancestors", func(t *testing.T) {
		// Create hierarchy: grandparent -> parent -> child
		grandparent := em.CreateEntity()
		parent := em.CreateEntity()
		child := em.CreateEntity()

		em.SetParent(parent, grandparent)
		_ = em.SetParent(child, parent)

		ancestors := em.GetAncestors(child)
		if len(ancestors) != 2 {
			t.Errorf("Child should have 2 ancestors, got %d", len(ancestors))
		}

		ancestorMap := make(map[EntityID]bool)
		for _, anc := range ancestors {
			ancestorMap[anc] = true
		}

		if !ancestorMap[parent] || !ancestorMap[grandparent] {
			t.Error("Ancestors should include both parent and grandparent")
		}
	})

	t.Run("TC030: Handle deep hierarchy (10 levels)", func(t *testing.T) {
		entities := make([]EntityID, 11) // 0-10 levels
		for i := range entities {
			entities[i] = em.CreateEntity()
		}

		// Create chain: 0 -> 1 -> 2 -> ... -> 10
		for i := 1; i < len(entities); i++ {
			em.SetParent(entities[i], entities[i-1])
		}

		ancestors := em.GetAncestors(entities[10])
		if len(ancestors) != 10 {
			t.Errorf("Deep child should have 10 ancestors, got %d", len(ancestors))
		}

		descendants := em.GetDescendants(entities[0])
		if len(descendants) != 10 {
			t.Errorf("Deep parent should have 10 descendants, got %d", len(descendants))
		}
	})

	t.Run("TC031: IsAncestor correctly identifies relationships", func(t *testing.T) {
		grandparent := em.CreateEntity()
		parent := em.CreateEntity()
		child := em.CreateEntity()

		em.SetParent(parent, grandparent)
		_ = em.SetParent(child, parent)

		if !em.IsAncestor(grandparent, child) {
			t.Error("Grandparent should be ancestor of child")
		}

		if !em.IsAncestor(parent, child) {
			t.Error("Parent should be ancestor of child")
		}

		if em.IsAncestor(child, grandparent) {
			t.Error("Child should not be ancestor of grandparent")
		}
	})
}

// TestEntityManager_RemoveFromParent tests parent relationship removal.
func TestEntityManager_RemoveFromParent(t *testing.T) {
	em := NewDefaultEntityManager()

	t.Run("TC032: Remove parent-child relationship", func(t *testing.T) {
		parent := em.CreateEntity()
		child := em.CreateEntity()

		_ = em.SetParent(child, parent)
		err := em.RemoveFromParent(child)
		if err != nil {
			t.Errorf("Should be able to remove parent relationship, got error: %v", err)
		}
	})

	t.Run("TC033: No parent after removal", func(t *testing.T) {
		parent := em.CreateEntity()
		child := em.CreateEntity()

		_ = em.SetParent(child, parent)
		em.RemoveFromParent(child)

		_, exists := em.GetParent(child)
		if exists {
			t.Error("Child should not have parent after removal")
		}
	})

	t.Run("TC034: Child removed from parent's children list", func(t *testing.T) {
		parent := em.CreateEntity()
		child := em.CreateEntity()

		_ = em.SetParent(child, parent)
		em.RemoveFromParent(child)

		children := em.GetChildren(parent)
		for _, c := range children {
			if c == child {
				t.Error("Child should be removed from parent's children list")
			}
		}
	})
}

// TestEntityManager_Tags tests entity tagging functionality.
func TestEntityManager_Tags(t *testing.T) {
	em := NewDefaultEntityManager()

	t.Run("TC035: Set entity tag", func(t *testing.T) {
		entity := em.CreateEntity()
		tag := "player"

		err := em.SetTag(entity, tag)
		if err != nil {
			t.Errorf("Should be able to set entity tag, got error: %v", err)
		}
	})

	t.Run("TC036: Get entity tag", func(t *testing.T) {
		entity := em.CreateEntity()
		expectedTag := "enemy"

		em.SetTag(entity, expectedTag)
		retrievedTag, exists := em.GetTag(entity)

		if !exists {
			t.Error("Should be able to get entity tag")
		}
		if retrievedTag != expectedTag {
			t.Errorf("Retrieved tag should be %s, got %s", expectedTag, retrievedTag)
		}
	})

	t.Run("TC037: Remove entity tag", func(t *testing.T) {
		entity := em.CreateEntity()
		em.SetTag(entity, "temporary")

		err := em.RemoveTag(entity)
		if err != nil {
			t.Errorf("Should be able to remove entity tag, got error: %v", err)
		}

		_, exists := em.GetTag(entity)
		if exists {
			t.Error("Tag should not exist after removal")
		}
	})

	t.Run("TC038: Find entities by tag", func(t *testing.T) {
		entity1 := em.CreateEntity()
		entity2 := em.CreateEntity()
		entity3 := em.CreateEntity()

		tag := "collectible"
		em.SetTag(entity1, tag)
		em.SetTag(entity2, tag)
		em.SetTag(entity3, "other")

		entities := em.FindByTag(tag)
		if len(entities) != 2 {
			t.Errorf("Should find 2 entities with tag %s, got %d", tag, len(entities))
		}

		entityMap := make(map[EntityID]bool)
		for _, e := range entities {
			entityMap[e] = true
		}

		if !entityMap[entity1] || !entityMap[entity2] {
			t.Error("Should find both entity1 and entity2 with the tag")
		}
		if entityMap[entity3] {
			t.Error("Should not find entity3 with different tag")
		}
	})

	t.Run("TC039: Empty result for non-existent tag", func(t *testing.T) {
		entities := em.FindByTag("nonexistent")
		if len(entities) != 0 {
			t.Errorf("Should return empty slice for non-existent tag, got %d entities", len(entities))
		}
	})

	t.Run("TC040: Get all tags", func(t *testing.T) {
		entity1 := em.CreateEntity()
		entity2 := em.CreateEntity()

		em.SetTag(entity1, "tag1")
		em.SetTag(entity2, "tag2")

		tags := em.GetAllTags()
		if len(tags) < 2 {
			t.Errorf("Should have at least 2 tags, got %d", len(tags))
		}

		tagMap := make(map[string]bool)
		for _, tag := range tags {
			tagMap[tag] = true
		}

		if !tagMap["tag1"] || !tagMap["tag2"] {
			t.Error("All tags should include both tag1 and tag2")
		}
	})
}

// TestEntityManager_Groups tests entity grouping functionality.
func TestEntityManager_Groups(t *testing.T) {
	em := NewDefaultEntityManager()

	t.Run("TC041: Create group", func(t *testing.T) {
		err := em.CreateGroup("enemies")
		if err != nil {
			t.Errorf("Should be able to create group, got error: %v", err)
		}
	})

	t.Run("TC042: Add entity to group", func(t *testing.T) {
		entity := em.CreateEntity()
		group := "players"

		em.CreateGroup(group)
		err := em.AddToGroup(entity, group)
		if err != nil {
			t.Errorf("Should be able to add entity to group, got error: %v", err)
		}
	})

	t.Run("TC043: Remove entity from group", func(t *testing.T) {
		entity := em.CreateEntity()
		group := "temporary"

		em.CreateGroup(group)
		em.AddToGroup(entity, group)

		err := em.RemoveFromGroup(entity, group)
		if err != nil {
			t.Errorf("Should be able to remove entity from group, got error: %v", err)
		}
	})

	t.Run("TC044: Get group entities", func(t *testing.T) {
		entity1 := em.CreateEntity()
		entity2 := em.CreateEntity()
		group := "items"

		em.CreateGroup(group)
		em.AddToGroup(entity1, group)
		em.AddToGroup(entity2, group)

		entities := em.GetGroup(group)
		if len(entities) != 2 {
			t.Errorf("Group should contain 2 entities, got %d", len(entities))
		}

		entityMap := make(map[EntityID]bool)
		for _, e := range entities {
			entityMap[e] = true
		}

		if !entityMap[entity1] || !entityMap[entity2] {
			t.Error("Group should contain both entity1 and entity2")
		}
	})

	t.Run("TC045: Get entity groups", func(t *testing.T) {
		entity := em.CreateEntity()
		group1 := "group1"
		group2 := "group2"

		em.CreateGroup(group1)
		em.CreateGroup(group2)
		em.AddToGroup(entity, group1)
		em.AddToGroup(entity, group2)

		groups := em.GetEntityGroups(entity)
		if len(groups) != 2 {
			t.Errorf("Entity should belong to 2 groups, got %d", len(groups))
		}

		groupMap := make(map[string]bool)
		for _, g := range groups {
			groupMap[g] = true
		}

		if !groupMap[group1] || !groupMap[group2] {
			t.Error("Entity should belong to both group1 and group2")
		}
	})

	t.Run("TC046: Destroy group", func(t *testing.T) {
		group := "destroyable"
		em.CreateGroup(group)

		err := em.DestroyGroup(group)
		if err != nil {
			t.Errorf("Should be able to destroy group, got error: %v", err)
		}

		entities := em.GetGroup(group)
		if len(entities) != 0 {
			t.Error("Destroyed group should return empty entity list")
		}
	})

	t.Run("TC047: Error on non-existent group operations", func(t *testing.T) {
		entity := em.CreateEntity()

		err1 := em.AddToGroup(entity, "nonexistent")
		if err1 == nil {
			t.Error("Adding to non-existent group should return error")
		}

		err2 := em.RemoveFromGroup(entity, "nonexistent")
		if err2 == nil {
			t.Error("Removing from non-existent group should return error")
		}

		err3 := em.DestroyGroup("nonexistent")
		if err3 == nil {
			t.Error("Destroying non-existent group should return error")
		}
	})
}

// TestEntityManager_CreateEntities tests batch entity creation.
func TestEntityManager_CreateEntities(t *testing.T) {
	em := NewDefaultEntityManager()

	t.Run("TC048: Create multiple entities at once", func(t *testing.T) {
		count := 10
		entities := em.CreateEntities(count)

		if len(entities) != count {
			t.Errorf("Should create %d entities, got %d", count, len(entities))
		}

		// Check all entities are unique
		entityMap := make(map[EntityID]bool)
		for _, entity := range entities {
			if entityMap[entity] {
				t.Errorf("Duplicate entity ID found: %d", entity)
			}
			entityMap[entity] = true
		}
	})

	t.Run("TC049: Correct number of entities created", func(t *testing.T) {
		initialCount := em.GetEntityCount()
		createCount := 5

		em.CreateEntities(createCount)
		newCount := em.GetEntityCount()

		if newCount != initialCount+createCount {
			t.Errorf("Entity count should increase by %d, expected %d, got %d", createCount, initialCount+createCount, newCount)
		}
	})

	t.Run("TC050: Large batch creation performance", func(t *testing.T) {
		count := 1000
		start := time.Now()

		entities := em.CreateEntities(count)
		duration := time.Since(start)

		if len(entities) != count {
			t.Errorf("Should create %d entities, got %d", count, len(entities))
		}

		// Target: 1000 entities in 16.67ms (60 FPS)
		if duration > 16670000 { // 16.67ms in nanoseconds
			t.Errorf("Batch creation too slow: %v, should be < 16.67ms", duration)
		}
	})
}

// TestEntityManager_DestroyEntities tests batch entity destruction.
func TestEntityManager_DestroyEntities(t *testing.T) {
	em := NewDefaultEntityManager()

	t.Run("TC051: Destroy multiple entities at once", func(t *testing.T) {
		entities := em.CreateEntities(5)

		err := em.DestroyEntities(entities)
		if err != nil {
			t.Errorf("Should be able to destroy multiple entities, got error: %v", err)
		}
	})

	t.Run("TC052: All entities invalid after batch destruction", func(t *testing.T) {
		entities := em.CreateEntities(5)
		em.DestroyEntities(entities)

		for _, entity := range entities {
			if em.IsValid(entity) {
				t.Errorf("Entity %d should be invalid after destruction", entity)
			}
		}
	})

	t.Run("TC053: Partial destruction with invalid entities", func(t *testing.T) {
		validEntities := em.CreateEntities(3)
		invalidEntity := EntityID(99999)

		mixedEntities := append(validEntities, invalidEntity)
		err := em.DestroyEntities(mixedEntities)

		// Should partially succeed (implementation dependent)
		// At minimum, should not panic
		_ = err // Error handling depends on implementation

		for _, entity := range validEntities {
			if em.IsValid(entity) {
				t.Errorf("Valid entity %d should be destroyed", entity)
			}
		}
	})

	t.Run("TC054: Large batch destruction performance", func(t *testing.T) {
		count := 1000
		entities := em.CreateEntities(count)

		start := time.Now()
		err := em.DestroyEntities(entities)
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Batch destruction should succeed, got error: %v", err)
		}

		// Target: 1000 entities in 16.67ms (60 FPS)
		if duration > 16670000 { // 16.67ms in nanoseconds
			t.Errorf("Batch destruction too slow: %v, should be < 16.67ms", duration)
		}
	})
}

// TestEntityManager_ValidateEntities tests batch entity validation.
func TestEntityManager_ValidateEntities(t *testing.T) {
	em := NewDefaultEntityManager()

	t.Run("TC055: Return only valid entities", func(t *testing.T) {
		validEntities := em.CreateEntities(3)
		invalidEntity := EntityID(99999)

		mixedEntities := append(validEntities, invalidEntity)
		validatedEntities := em.ValidateEntities(mixedEntities)

		if len(validatedEntities) != len(validEntities) {
			t.Errorf("Should return %d valid entities, got %d", len(validEntities), len(validatedEntities))
		}

		validEntityMap := make(map[EntityID]bool)
		for _, entity := range validEntities {
			validEntityMap[entity] = true
		}

		for _, entity := range validatedEntities {
			if !validEntityMap[entity] {
				t.Errorf("Unexpected entity in validated list: %d", entity)
			}
		}
	})

	t.Run("TC056: Filter out invalid entities", func(t *testing.T) {
		validEntities := em.CreateEntities(2)
		invalidEntities := []EntityID{99998, 99999}

		mixedEntities := append(validEntities, invalidEntities...)
		validatedEntities := em.ValidateEntities(mixedEntities)

		for _, validated := range validatedEntities {
			for _, invalid := range invalidEntities {
				if validated == invalid {
					t.Errorf("Invalid entity %d should not be in validated list", invalid)
				}
			}
		}
	})

	t.Run("TC057: Empty input returns empty output", func(t *testing.T) {
		validatedEntities := em.ValidateEntities([]EntityID{})
		if len(validatedEntities) != 0 {
			t.Errorf("Empty input should return empty output, got %d entities", len(validatedEntities))
		}
	})
}

// TestEntityManager_Concurrent tests thread safety.
func TestEntityManager_Concurrent(t *testing.T) {
	em := NewDefaultEntityManager()

	t.Run("TC084: Concurrent entity creation", func(t *testing.T) {
		const goroutines = 10
		const entitiesPerGoroutine = 100

		var wg sync.WaitGroup
		entityChannels := make([]chan EntityID, goroutines)

		// Start concurrent entity creation
		for i := 0; i < goroutines; i++ {
			entityChannels[i] = make(chan EntityID, entitiesPerGoroutine)
			wg.Add(1)

			go func(ch chan EntityID) {
				defer wg.Done()
				defer close(ch)

				for j := 0; j < entitiesPerGoroutine; j++ {
					entity := em.CreateEntity()
					ch <- entity
				}
			}(entityChannels[i])
		}

		wg.Wait()

		// Collect all created entities
		allEntities := make(map[EntityID]bool)
		totalEntities := 0

		for _, ch := range entityChannels {
			for entity := range ch {
				if allEntities[entity] {
					t.Errorf("Duplicate entity ID created concurrently: %d", entity)
				}
				allEntities[entity] = true
				totalEntities++
			}
		}

		expectedTotal := goroutines * entitiesPerGoroutine
		if totalEntities != expectedTotal {
			t.Errorf("Expected %d total entities, got %d", expectedTotal, totalEntities)
		}
	})

	t.Run("TC085: Concurrent read operations", func(t *testing.T) {
		// Create some entities first
		entities := em.CreateEntities(100)

		const readers = 10
		var wg sync.WaitGroup

		for i := 0; i < readers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				// Perform various read operations
				for j := 0; j < 100; j++ {
					entity := entities[j%len(entities)]
					em.IsValid(entity)
					em.GetEntityCount()
					em.GetActiveEntities()
				}
			}()
		}

		wg.Wait()
		// If we get here without deadlock, the test passes
	})

	t.Run("TC086: Mixed read/write operations", func(t *testing.T) {
		const workers = 5
		var wg sync.WaitGroup

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				// Each worker performs mixed operations
				for j := 0; j < 50; j++ {
					if j%2 == 0 {
						// Write operations
						entity := em.CreateEntity()
						em.SetTag(entity, "worker")
					} else {
						// Read operations
						em.GetEntityCount()
						em.FindByTag("worker")
					}
				}
			}(i)
		}

		wg.Wait()
		// If we get here without race conditions, the test passes
	})

	t.Run("TC087: Large scale concurrent operations", func(t *testing.T) {
		const workers = 10
		const operationsPerWorker = 1000

		start := time.Now()
		var wg sync.WaitGroup

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for j := 0; j < operationsPerWorker; j++ {
					entity := em.CreateEntity()
					em.IsValid(entity)
					_ = em.DestroyEntity(entity)
				}
			}()
		}

		wg.Wait()
		duration := time.Since(start)

		// Should complete without data races or excessive time
		if duration > 5*time.Second {
			t.Errorf("Concurrent operations took too long: %v", duration)
		}
	})
}

// TestEntityManager_Performance benchmarks for performance validation.
func TestEntityManager_Performance(t *testing.T) {
	em := NewDefaultEntityManager()

	t.Run("PC001: Creation performance < 16.67ms for 1000 entities", func(t *testing.T) {
		start := time.Now()
		entities := em.CreateEntities(1000)
		duration := time.Since(start)

		if len(entities) != 1000 {
			t.Errorf("Should create 1000 entities, got %d", len(entities))
		}

		targetDuration := 16670000 // 16.67ms in nanoseconds
		if duration > time.Duration(targetDuration) {
			t.Errorf("Creation performance too slow: %v, target: 16.67ms", duration)
		}

		t.Logf("Created 1000 entities in %v", duration)
	})

	t.Run("PC002: Destruction performance < 16.67ms for 1000 entities", func(t *testing.T) {
		entities := em.CreateEntities(1000)

		start := time.Now()
		err := em.DestroyEntities(entities)
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Destruction should succeed, got error: %v", err)
		}

		targetDuration := 16670000 // 16.67ms in nanoseconds
		if duration > time.Duration(targetDuration) {
			t.Errorf("Destruction performance too slow: %v, target: 16.67ms", duration)
		}

		t.Logf("Destroyed 1000 entities in %v", duration)
	})

	t.Run("PC003: Memory usage < 100B per entity", func(t *testing.T) {
		initialMemory := em.GetMemoryUsage()
		entities := em.CreateEntities(1000)
		finalMemory := em.GetMemoryUsage()

		memoryPerEntity := (finalMemory - initialMemory) / int64(len(entities))

		if memoryPerEntity > 100 {
			t.Errorf("Memory usage per entity too high: %d bytes, target: <100B", memoryPerEntity)
		}

		t.Logf("Memory usage per entity: %d bytes", memoryPerEntity)
	})
}

// Helper function to run all entity manager tests.
func TestEntityManager_AllTests(t *testing.T) {
	// This is a placeholder to ensure all tests are discovered.
	// Individual test functions above will be run by go test.
	t.Log("EntityManager test suite loaded")
}
