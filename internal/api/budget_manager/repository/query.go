package budgetRepository

const (
	queryCreateTransaction = `
		INSERT INTO budget_transactions (
			id,
			user_id,
			title,
			description,
			nominal,
			type,
			category,
			audio_link,
			created_at,
			updated_at
		) VALUES (
			:id,
			:user_id,
			:title,
			:description,
			:nominal,
			:type,
			:category,
			:audio_link,
			:created_at,
			:updated_at
		)
	`

	queryGetAllTransactions = `
		SELECT
			id,
			user_id,
			title,
			description,
			nominal,
			type,
			category,
			audio_link,
			created_at,
			updated_at
		FROM budget_transactions
		WHERE user_id = :user_id
		ORDER BY created_at DESC
	`

	queryGetCurrentWeekTransactions = `
		SELECT
			id,
			user_id,
			title,
			description,
			nominal,
			type,
			category,
			audio_link,
			created_at,
			updated_at
		FROM budget_transactions
		WHERE 
			user_id = :user_id
			AND created_at >= date_trunc('week', CURRENT_DATE)
			AND created_at < date_trunc('week', CURRENT_DATE) + interval '1 week'
		ORDER BY created_at DESC
	`

	queryGetCurrentMonthTransactions = `
		SELECT
			id,
			user_id,
			title,
			description,
			nominal,
			type,
			category,
			audio_link,
			created_at,
			updated_at
		FROM budget_transactions
		WHERE 
			user_id = :user_id
			AND created_at >= date_trunc('month', CURRENT_DATE)
			AND created_at < date_trunc('month', CURRENT_DATE) + interval '1 month'
		ORDER BY created_at DESC
	`

	queryGetTransactionById = `
		SELECT
			id,
			user_id,
			title,
			description,
			nominal,
			type,
			category,
			audio_link,
			created_at,
			updated_at
		FROM budget_transactions
		WHERE id = :id
	`

	queryGetTransactionsByUserID = `
		SELECT
			id,
			user_id,
			title,
			description,
			nominal,
			type,
			category,
			audio_link,
			created_at,
			updated_at
		FROM budget_transactions
		WHERE user_id = :user_id
		ORDER BY created_at DESC
	`

	queryUpdateTransaction = `
		UPDATE budget_transactions
		SET
			title = :title,
			description = :description,
			nominal = :nominal,
			type = :type,
			category = :category,
			audio_link = :audio_link,
			updated_at = :updated_at
		WHERE id = :id
	`

	queryDeleteTransaction = `
		DELETE FROM budget_transactions
		WHERE id = :id
	`

	queryGetTransactionsByTypeAndCategory = `
		SELECT
			id,
			user_id,
			title,
			description,
			nominal,
			type,
			category,
			audio_link,
			created_at,
			updated_at
		FROM budget_transactions
		WHERE 
			user_id = :user_id
			AND type = :type
			AND category = :category
		ORDER BY created_at DESC
	`
)
