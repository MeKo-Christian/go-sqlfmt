SELECT
  jt.product_id,
  jt.product_name,
  jt.price,
  jt.category,
  jt.tags,
  jt.specifications
FROM
  products p,
  JSON_TABLE(
    p.product_data,
    '$' COLUMNS(
      product_id INT PATH '$.id',
      product_name VARCHAR(100) PATH '$.name',
      price DECIMAL(10, 2) PATH '$.pricing.base_price',
      category VARCHAR(50) PATH '$.category',
      tags JSON PATH '$.tags',
      specifications JSON PATH '$.specifications'
    )
  ) AS jt
WHERE
  p.active = 1;