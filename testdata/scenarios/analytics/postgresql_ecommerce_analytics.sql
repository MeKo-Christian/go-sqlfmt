-- PostgreSQL Analytics: E-commerce sales and customer behavior analysis
-- Complex analytical queries with window functions, CTEs, and aggregations

-- Monthly sales performance analysis
WITH monthly_sales AS (
    SELECT
        DATE_TRUNC('month', o.order_date) as sales_month,
        COUNT(DISTINCT o.id) as total_orders,
        COUNT(DISTINCT o.customer_id) as unique_customers,
        SUM(o.total_amount) as total_revenue,
        AVG(o.total_amount) as avg_order_value,
        SUM(o.total_amount) / COUNT(DISTINCT o.customer_id) as revenue_per_customer,
        COUNT(CASE WHEN o.status = 'completed' THEN 1 END) as completed_orders,
        COUNT(CASE WHEN o.status = 'cancelled' THEN 1 END) as cancelled_orders
    FROM orders o
    WHERE o.order_date >= CURRENT_DATE - INTERVAL '12 months'
    GROUP BY DATE_TRUNC('month', o.order_date)
),
monthly_growth AS (
    SELECT
        sales_month,
        total_orders,
        unique_customers,
        total_revenue,
        avg_order_value,
        revenue_per_customer,
        completed_orders,
        cancelled_orders,
        -- Calculate month-over-month growth
        ROUND(
            (total_revenue - LAG(total_revenue) OVER (ORDER BY sales_month)) /
            NULLIF(LAG(total_revenue) OVER (ORDER BY sales_month), 0) * 100,
            2
        ) as revenue_growth_pct,
        ROUND(
            (total_orders - LAG(total_orders) OVER (ORDER BY sales_month)) /
            NULLIF(LAG(total_orders) OVER (ORDER BY sales_month), 0) * 100,
            2
        ) as orders_growth_pct,
        -- Calculate conversion rate
        ROUND(completed_orders::numeric / NULLIF(total_orders, 0) * 100, 2) as completion_rate
    FROM monthly_sales
)
SELECT
    TO_CHAR(sales_month, 'Mon YYYY') as month,
    total_orders,
    unique_customers,
    ROUND(total_revenue, 2) as total_revenue,
    ROUND(avg_order_value, 2) as avg_order_value,
    ROUND(revenue_per_customer, 2) as revenue_per_customer,
    revenue_growth_pct,
    orders_growth_pct,
    completion_rate
FROM monthly_growth
ORDER BY sales_month DESC;

-- Customer lifetime value and segmentation analysis
WITH customer_orders AS (
    SELECT
        c.id as customer_id,
        c.email,
        c.created_at as customer_since,
        COUNT(o.id) as total_orders,
        SUM(o.total_amount) as total_spent,
        AVG(o.total_amount) as avg_order_value,
        MAX(o.order_date) as last_order_date,
        MIN(o.order_date) as first_order_date,
        EXTRACT(EPOCH FROM (MAX(o.order_date) - MIN(o.order_date))) / 86400 as customer_lifespan_days,
        EXTRACT(EPOCH FROM (CURRENT_DATE - MAX(o.order_date))) / 86400 as days_since_last_order
    FROM customers c
    LEFT JOIN orders o ON c.id = o.customer_id AND o.status = 'completed'
    WHERE c.is_active = TRUE
    GROUP BY c.id, c.email, c.created_at
),
customer_segments AS (
    SELECT
        customer_id,
        email,
        customer_since,
        total_orders,
        total_spent,
        avg_order_value,
        last_order_date,
        first_order_date,
        customer_lifespan_days,
        days_since_last_order,
        -- Calculate customer lifetime value (CLV)
        CASE
            WHEN total_orders > 0 THEN
                total_spent / GREATEST(total_orders, 1) * GREATEST(total_orders, 1) -- Simplified CLV
            ELSE 0
        END as estimated_clv,
        -- Segment customers based on spending and recency
        CASE
            WHEN total_spent >= 1000 AND days_since_last_order <= 30 THEN 'High-Value Active'
            WHEN total_spent >= 500 AND days_since_last_order <= 90 THEN 'Medium-Value Recent'
            WHEN total_spent >= 100 THEN 'Low-Value Regular'
            WHEN total_orders > 0 THEN 'One-Time Buyer'
            ELSE 'Prospect'
        END as customer_segment
    FROM customer_orders
),
segment_summary AS (
    SELECT
        customer_segment,
        COUNT(*) as customer_count,
        ROUND(AVG(total_spent), 2) as avg_total_spent,
        ROUND(AVG(total_orders), 2) as avg_orders_per_customer,
        ROUND(AVG(avg_order_value), 2) as avg_order_value,
        ROUND(AVG(days_since_last_order), 1) as avg_days_since_last_order,
        ROUND(SUM(total_spent), 2) as segment_total_revenue,
        ROUND(AVG(estimated_clv), 2) as avg_estimated_clv
    FROM customer_segments
    GROUP BY customer_segment
)
SELECT
    customer_segment,
    customer_count,
    ROUND(customer_count::numeric / SUM(customer_count) OVER () * 100, 1) as segment_percentage,
    avg_total_spent,
    avg_orders_per_customer,
    avg_order_value,
    avg_days_since_last_order,
    segment_total_revenue,
    ROUND(segment_total_revenue / SUM(segment_total_revenue) OVER () * 100, 1) as revenue_percentage,
    avg_estimated_clv
FROM segment_summary
ORDER BY segment_total_revenue DESC;

-- Product performance analysis with cohort analysis
WITH product_sales AS (
    SELECT
        p.id as product_id,
        p.name as product_name,
        c.name as category_name,
        DATE_TRUNC('month', o.order_date) as sale_month,
        SUM(oi.quantity) as units_sold,
        SUM(oi.total_price) as revenue,
        COUNT(DISTINCT o.customer_id) as unique_customers,
        AVG(oi.unit_price) as avg_selling_price
    FROM products p
    JOIN product_categories c ON p.category_id = c.id
    JOIN order_items oi ON p.id = oi.product_id
    JOIN orders o ON oi.order_id = o.id
    WHERE o.status = 'completed'
      AND o.order_date >= CURRENT_DATE - INTERVAL '6 months'
    GROUP BY p.id, p.name, c.name, DATE_TRUNC('month', o.order_date)
),
product_metrics AS (
    SELECT
        product_id,
        product_name,
        category_name,
        sale_month,
        units_sold,
        revenue,
        unique_customers,
        avg_selling_price,
        -- Calculate growth metrics
        ROUND(
            (revenue - LAG(revenue) OVER (PARTITION BY product_id ORDER BY sale_month)) /
            NULLIF(LAG(revenue) OVER (PARTITION BY product_id ORDER BY sale_month), 0) * 100,
            2
        ) as revenue_growth_pct,
        -- Calculate market share within category
        ROUND(
            revenue / SUM(revenue) OVER (PARTITION BY category_name, sale_month) * 100,
            2
        ) as category_market_share,
        -- Rank products within category by revenue
        ROW_NUMBER() OVER (PARTITION BY category_name, sale_month ORDER BY revenue DESC) as category_rank
    FROM product_sales
),
top_products AS (
    SELECT
        product_name,
        category_name,
        SUM(units_sold) as total_units_sold,
        SUM(revenue) as total_revenue,
        AVG(avg_selling_price) as avg_price,
        COUNT(DISTINCT sale_month) as active_months,
        ROUND(AVG(revenue_growth_pct), 2) as avg_monthly_growth,
        ROUND(AVG(category_market_share), 2) as avg_market_share
    FROM product_metrics
    GROUP BY product_id, product_name, category_name
    HAVING SUM(revenue) > 0
    ORDER BY total_revenue DESC
    LIMIT 20
)
SELECT
    ROW_NUMBER() OVER (ORDER BY total_revenue DESC) as rank,
    product_name,
    category_name,
    total_units_sold,
    ROUND(total_revenue, 2) as total_revenue,
    ROUND(avg_price, 2) as avg_price,
    active_months,
    avg_monthly_growth,
    avg_market_share
FROM top_products;

-- Complex inventory turnover and stock analysis
WITH inventory_movements AS (
    SELECT
        p.id as product_id,
        p.name as product_name,
        p.inventory_quantity as current_stock,
        p.cost_price,
        COALESCE(SUM(oi.quantity), 0) as units_sold_last_30_days,
        COALESCE(SUM(oi.total_price), 0) as revenue_last_30_days,
        MAX(o.order_date) as last_sale_date
    FROM products p
    LEFT JOIN order_items oi ON p.id = oi.product_id
    LEFT JOIN orders o ON oi.order_id = o.id
        AND o.status = 'completed'
        AND o.order_date >= CURRENT_DATE - INTERVAL '30 days'
    GROUP BY p.id, p.name, p.inventory_quantity, p.cost_price
),
inventory_analysis AS (
    SELECT
        product_id,
        product_name,
        current_stock,
        cost_price,
        units_sold_last_30_days,
        revenue_last_30_days,
        last_sale_date,
        -- Calculate inventory turnover (units sold / average inventory)
        CASE
            WHEN current_stock > 0 THEN
                ROUND(units_sold_last_30_days::numeric / (current_stock / 2), 2)
            ELSE 0
        END as inventory_turnover_ratio,
        -- Calculate days of inventory
        CASE
            WHEN units_sold_last_30_days > 0 THEN
                ROUND(current_stock::numeric / (units_sold_last_30_days / 30), 1)
            ELSE NULL
        END as days_of_inventory,
        -- Calculate stock status
        CASE
            WHEN current_stock = 0 THEN 'Out of Stock'
            WHEN current_stock <= 5 THEN 'Low Stock'
            WHEN days_of_inventory > 90 THEN 'Overstocked'
            WHEN last_sale_date < CURRENT_DATE - INTERVAL '30 days' THEN 'Slow Moving'
            ELSE 'Normal'
        END as stock_status,
        -- Calculate gross margin
        CASE
            WHEN cost_price > 0 THEN
                ROUND((revenue_last_30_days - (units_sold_last_30_days * cost_price)) / revenue_last_30_days * 100, 2)
            ELSE NULL
        END as gross_margin_pct
    FROM inventory_movements
),
stock_summary AS (
    SELECT
        stock_status,
        COUNT(*) as product_count,
        SUM(current_stock) as total_stock_units,
        ROUND(AVG(days_of_inventory), 1) as avg_days_of_inventory,
        ROUND(AVG(inventory_turnover_ratio), 2) as avg_turnover_ratio,
        ROUND(AVG(gross_margin_pct), 2) as avg_gross_margin
    FROM inventory_analysis
    GROUP BY stock_status
)
SELECT
    stock_status,
    product_count,
    ROUND(product_count::numeric / SUM(product_count) OVER () * 100, 1) as percentage_of_products,
    total_stock_units,
    ROUND(total_stock_units::numeric / SUM(total_stock_units) OVER () * 100, 1) as percentage_of_stock,
    avg_days_of_inventory,
    avg_turnover_ratio,
    avg_gross_margin
FROM stock_summary
ORDER BY
    CASE stock_status
        WHEN 'Out of Stock' THEN 1
        WHEN 'Low Stock' THEN 2
        WHEN 'Slow Moving' THEN 3
        WHEN 'Overstocked' THEN 4
        WHEN 'Normal' THEN 5
    END;