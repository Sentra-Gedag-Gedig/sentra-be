package sentrapayRepository

const (
	queryCreateWallet = `
		INSERT INTO wallets (
			id,
			user_id,
			balance,
			created_at,
			updated_at
		) VALUES (
			:id,
			:user_id,
			:balance,
			:created_at,
			:updated_at
		)
	`

	queryGetWallet = `
		SELECT
			id,
			user_id,
			balance,
			created_at,
			updated_at
		FROM wallets
		WHERE user_id = :user_id
	`

	queryUpdateWalletBalance = `
		UPDATE wallets
		SET
			balance = :balance,
			updated_at = :updated_at
		WHERE user_id = :user_id
	`

	queryCreateTransaction = `
		INSERT INTO wallet_transactions (
			id,
			user_id,
			amount,
			type,
			reference_no,
			payment_method,
			status,
			bank_account,
			bank_name,
			description,
			created_at,
			updated_at
		) VALUES (
			:id,
			:user_id,
			:amount,
			:type,
			:reference_no,
			:payment_method,
			:status,
			:bank_account,
			:bank_name,
			:description,
			:created_at,
			:updated_at
		)
	`

	queryGetTransactionByID = `
		SELECT
			id,
			user_id,
			amount,
			type,
			reference_no,
			payment_method,
			status,
			bank_account,
			bank_name,
			description,
			created_at,
			updated_at
		FROM wallet_transactions
		WHERE id = :id
	`

	queryGetTransactionByReferenceNo = `
		SELECT
			id,
			user_id,
			amount,
			type,
			reference_no,
			payment_method,
			status,
			bank_account,
			bank_name,
			description,
			created_at,
			updated_at
		FROM wallet_transactions
		WHERE reference_no = :reference_no
	`

	queryUpdateTransactionStatus = `
		UPDATE wallet_transactions
		SET
			status = :status,
			updated_at = :updated_at
		WHERE reference_no = :reference_no
	`

	queryGetTransactionsByUserID = `
		SELECT
			id,
			user_id,
			amount,
			type,
			reference_no,
			payment_method,
			status,
			bank_account,
			bank_name,
			description,
			created_at,
			updated_at
		FROM wallet_transactions
		WHERE user_id = :user_id
		ORDER BY created_at DESC
		LIMIT :limit OFFSET :offset
	`

	queryCountTransactionsByUserID = `
		SELECT COUNT(*)
		FROM wallet_transactions
		WHERE user_id = :user_id
	`
)
