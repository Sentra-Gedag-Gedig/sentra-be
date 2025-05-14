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

	queryCreateQRisTransaction = `
		INSERT INTO qris_transactions (
			id,
			user_id,
			recipient_id,
			amount,
			fee,
			total_amount,
			type,
			reference_no,
			status,
			description,
			qris_code,
			payment_datetime,
			created_at,
			updated_at
		) VALUES (
			:id,
			:user_id,
			:recipient_id,
			:amount,
			:fee,
			:total_amount,
			:type,
			:reference_no,
			:status,
			:description,
			:qris_code,
			:payment_datetime,
			:created_at,
			:updated_at
		)
	`

	queryGetQRisTransactionByID = `
		SELECT
			qt.id,
			qt.user_id,
			qt.recipient_id,
			qt.amount,
			qt.fee,
			qt.total_amount,
			qt.type,
			qt.reference_no,
			qt.status,
			qt.description,
			qt.qris_code,
			qt.payment_datetime,
			qt.created_at,
			qt.updated_at,
			u.name as recipient_name
		FROM qris_transactions qt
		LEFT JOIN users u ON qt.recipient_id = u.id
		WHERE qt.id = :id
	`

	queryGetQRisTransactionByReferenceNo = `
		SELECT
			qt.id,
			qt.user_id,
			qt.recipient_id,
			qt.amount,
			qt.fee,
			qt.total_amount,
			qt.type,
			qt.reference_no,
			qt.status,
			qt.description,
			qt.qris_code,
			qt.payment_datetime,
			qt.created_at,
			qt.updated_at,
			u.name as recipient_name
		FROM qris_transactions qt
		LEFT JOIN users u ON qt.recipient_id = u.id
		WHERE qt.reference_no = :reference_no
	`

	queryUpdateQRisTransactionStatus = `
		UPDATE qris_transactions
		SET
			status = :status,
			updated_at = :updated_at,
			payment_datetime = :payment_datetime
		WHERE reference_no = :reference_no
	`

	queryGetQRisTransactionsByUserID = `
		SELECT
			qt.id,
			qt.user_id,
			qt.recipient_id,
			qt.amount,
			qt.fee,
			qt.total_amount,
			qt.type,
			qt.reference_no,
			qt.status,
			qt.description,
			qt.qris_code,
			qt.payment_datetime,
			qt.created_at,
			qt.updated_at,
			u.name as recipient_name
		FROM qris_transactions qt
		LEFT JOIN users u ON qt.recipient_id = u.id
		WHERE qt.user_id = :user_id
		  AND (:start_date IS NULL OR qt.created_at >= :start_date)
		  AND (:end_date IS NULL OR qt.created_at <= :end_date)
		  AND (:type IS NULL OR qt.type = :type)
		  AND (:status IS NULL OR qt.status = :status)
		ORDER BY qt.created_at DESC
		LIMIT :limit OFFSET :offset
	`

	queryCountQRisTransactionsByUserID = `
		SELECT COUNT(*)
		FROM qris_transactions qt
		WHERE qt.user_id = :user_id
		  AND (:start_date IS NULL OR qt.created_at >= :start_date)
		  AND (:end_date IS NULL OR qt.created_at <= :end_date)
		  AND (:type IS NULL OR qt.type = :type)
		  AND (:status IS NULL OR qt.status = :status)
	`

	queryGetQRisTransactionByQRisCode = `
		SELECT
			qt.id,
			qt.user_id,
			qt.recipient_id,
			qt.amount,
			qt.fee,
			qt.total_amount,
			qt.type,
			qt.reference_no,
			qt.status,
			qt.description,
			qt.qris_code,
			qt.payment_datetime,
			qt.created_at,
			qt.updated_at,
			u.name as recipient_name
		FROM qris_transactions qt
		LEFT JOIN users u ON qt.recipient_id = u.id
		WHERE qt.qris_code = :qris_code
		  AND qt.status = 'pending'
	`

	queryGetActiveQRisCode = `
		SELECT qris_code
		FROM qris_transactions
		WHERE user_id = :user_id
		  AND status = 'pending'
		  AND created_at > :cutoff_time
		LIMIT 1
	`
)
