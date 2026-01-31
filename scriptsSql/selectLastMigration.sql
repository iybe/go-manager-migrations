select *
from (
	select migration_name
	from t_migrations
	group by migration_name
	order by migration_name desc
)
as x
limit 1;