package sqlfmt

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestMemoryUsage_ProfileAllocation tests memory allocation patterns during formatting
func TestMemoryUsage_ProfileAllocation(t *testing.T) {
	// Force garbage collection before starting
	runtime.GC()
	runtime.GC() // Run twice to ensure cleanup

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Test with a large, complex query
	largeQuery := `WITH RECURSIVE employee_hierarchy AS (
		SELECT
			emp_id,
			name,
			manager_id,
			department,
			salary,
			hire_date,
			0 as level,
			ARRAY[emp_id] as path
		FROM employees
		WHERE manager_id IS NULL
		UNION ALL
		SELECT
			e.emp_id,
			e.name,
			e.manager_id,
			e.department,
			e.salary,
			e.hire_date,
			eh.level + 1,
			eh.path || e.emp_id
		FROM employees e
		JOIN employee_hierarchy eh ON e.manager_id = eh.emp_id
		WHERE eh.level < 10
	),
	department_stats AS (
		SELECT
			department,
			COUNT(*) as total_employees,
			AVG(salary) as avg_salary,
			MIN(salary) as min_salary,
			MAX(salary) as max_salary,
			SUM(salary) as total_salary,
			STDDEV(salary) as salary_stddev,
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY salary) as median_salary,
			STRING_AGG(name, ', ' ORDER BY salary DESC) as top_earners
		FROM employee_hierarchy
		GROUP BY department
	)
	SELECT
		eh.name,
		eh.department,
		eh.level,
		eh.salary,
		ds.avg_salary,
		ds.median_salary,
		RANK() OVER (PARTITION BY eh.department ORDER BY eh.salary DESC) as dept_salary_rank,
		DENSE_RANK() OVER (PARTITION BY eh.department ORDER BY eh.salary DESC) as dept_dense_rank,
		ROW_NUMBER() OVER (PARTITION BY eh.department ORDER BY eh.salary DESC) as dept_row_num,
		LAG(eh.salary) OVER (PARTITION BY eh.department ORDER BY eh.salary DESC) as next_lower_salary,
		LEAD(eh.salary) OVER (PARTITION BY eh.department ORDER BY eh.salary DESC) as next_higher_salary,
		COUNT(*) OVER (PARTITION BY eh.department) as dept_size,
		SUM(eh.salary) OVER (PARTITION BY eh.department) as dept_total_salary
	FROM employee_hierarchy eh
	JOIN department_stats ds ON eh.department = ds.department
	WHERE eh.salary > ds.median_salary
	ORDER BY eh.department, eh.salary DESC, eh.name`

	// Run formatting multiple times to get stable measurements
	for i := 0; i < 100; i++ {
		Format(largeQuery)
	}

	runtime.ReadMemStats(&m2)

	// Calculate allocations
	allocations := m2.TotalAlloc - m1.TotalAlloc
	t.Logf("Total allocations during 100 format operations: %d bytes (%.2f MB)", allocations, float64(allocations)/(1024*1024))

	// Ensure we didn't have excessive allocations (arbitrary threshold: < 200MB for 100 operations of complex SQL)
	if allocations > 200*1024*1024 {
		t.Errorf("Excessive memory allocations: %d bytes (> 200MB)", allocations)
	}
}

// TestMemoryUsage_NoLeaks tests for memory leaks with repeated formatting operations
func TestMemoryUsage_NoLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	// Force initial GC
	runtime.GC()
	runtime.GC()

	var initialStats runtime.MemStats
	runtime.ReadMemStats(&initialStats)

	query := `SELECT
		u.id,
		u.username,
		u.email,
		p.first_name,
		p.last_name,
		CASE
			WHEN u.status = 'active' THEN 'Active User'
			WHEN u.status = 'inactive' THEN 'Inactive User'
			ELSE 'Unknown Status'
		END as status_description
	FROM users u
	LEFT JOIN profiles p ON u.id = p.user_id
	WHERE u.active = true
	AND u.email_verified = true
	AND p.created_at > '2023-01-01'
	ORDER BY u.username ASC, p.created_at DESC
	LIMIT 100`

	// Run many iterations to detect potential leaks
	const iterations = 10000
	for i := 0; i < iterations; i++ {
		Format(query)
		if i%1000 == 0 {
			runtime.GC() // Periodic GC to clean up
		}
	}

	// Force final GC
	runtime.GC()
	runtime.GC()

	var finalStats runtime.MemStats
	runtime.ReadMemStats(&finalStats)

	// Check heap allocations - should not grow significantly
	heapGrowth := finalStats.HeapAlloc - initialStats.HeapAlloc
	t.Logf("Heap growth after %d operations: %d bytes (%.2f MB)", iterations, heapGrowth, float64(heapGrowth)/(1024*1024))

	// Allow some growth but not excessive (arbitrary threshold: < 10MB growth)
	if heapGrowth > 10*1024*1024 {
		t.Errorf("Potential memory leak detected: heap grew by %d bytes (> 10MB)", heapGrowth)
	}
}

// TestConcurrentFormatting tests concurrent formatting operations for thread safety and performance
func TestConcurrentFormatting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent formatting test in short mode")
	}

	const numGoroutines = 10
	const operationsPerGoroutine = 1000

	queries := []string{
		"SELECT id, name FROM users WHERE active = true",
		`SELECT u.id, u.username, u.email, p.first_name, p.last_name
		 FROM users u LEFT JOIN profiles p ON u.id = p.user_id
		 WHERE u.active = true ORDER BY u.username`,
		`WITH RECURSIVE t(n) AS (VALUES (1) UNION ALL SELECT n+1 FROM t WHERE n < 100)
		 SELECT n FROM t`,
	}

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				query := queries[j%len(queries)]
				result := Format(query)
				if result == "" {
					t.Errorf("Goroutine %d: formatting failed for query %d", goroutineID, j)
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	totalOperations := numGoroutines * operationsPerGoroutine
	opsPerSecond := float64(totalOperations) / duration.Seconds()

	t.Logf("Concurrent formatting: %d goroutines × %d operations = %d total operations",
		numGoroutines, operationsPerGoroutine, totalOperations)
	t.Logf("Total time: %v", duration)
	t.Logf("Operations per second: %.2f", opsPerSecond)

	// Ensure reasonable performance (arbitrary threshold: > 500 ops/sec)
	if opsPerSecond < 500 {
		t.Errorf("Performance too low: %.2f operations/second (< 500)", opsPerSecond)
	}
}

// BenchmarkMemoryUsage tracks memory usage during benchmark operations
func BenchmarkMemoryUsage(b *testing.B) {
	query := `SELECT
		u.id,
		u.username,
		u.email,
		p.first_name,
		p.last_name,
		o.total_amount,
		o.created_at
	FROM users u
	LEFT JOIN profiles p ON u.id = p.user_id
	LEFT JOIN orders o ON u.id = o.user_id
	WHERE u.active = true
	AND o.status = 'completed'
	ORDER BY u.username, o.created_at DESC`

	// Reset memory stats
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		Format(query)
	}
}

// BenchmarkConcurrentFormatting benchmarks concurrent formatting operations
func BenchmarkConcurrentFormatting(b *testing.B) {
	query := `SELECT
		u.id,
		u.username,
		u.email,
		p.first_name,
		p.last_name
	FROM users u
	LEFT JOIN profiles p ON u.id = p.user_id
	WHERE u.active = true
	ORDER BY u.username`

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Format(query)
		}
	})
}
