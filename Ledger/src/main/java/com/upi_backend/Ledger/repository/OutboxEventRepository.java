package com.upi_backend.Ledger.repository;

import com.upi_backend.Ledger.model.OutboxEvent;
import org.springframework.data.jpa.repository.JpaRepository;

public interface OutboxEventRepository extends JpaRepository<OutboxEvent, String> {
}