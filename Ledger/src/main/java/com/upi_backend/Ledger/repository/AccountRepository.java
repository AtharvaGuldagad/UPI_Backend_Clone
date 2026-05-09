package com.upi_backend.Ledger.repository;

import com.upi_backend.Ledger.model.Account;
import org.springframework.data.jpa.repository.JpaRepository;

public interface AccountRepository extends JpaRepository<Account, String> {
}