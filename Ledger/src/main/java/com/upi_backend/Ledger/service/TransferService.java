package com.upi_backend.Ledger.service;

import com.upi_backend.Ledger.model.Account;
import com.upi_backend.Ledger.model.OutboxEvent;
import com.upi_backend.Ledger.repository.AccountRepository;
import com.upi_backend.Ledger.repository.OutboxEventRepository;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import java.math.BigDecimal;

@Service
public class TransferService {

    private final AccountRepository accountRepository;
    private final OutboxEventRepository outboxEventRepository;

    public TransferService(AccountRepository accountRepository, OutboxEventRepository outboxEventRepository) {
        this.accountRepository = accountRepository;
        this.outboxEventRepository = outboxEventRepository;
    }

    @Transactional
    public void transferMoney(String fromAccountId, String toAccountId, BigDecimal amount) {
        Account fromAccount = accountRepository.findById(fromAccountId)
                .orElseThrow(() -> new RuntimeException("Sender account not found"));
        
        Account toAccount = accountRepository.findById(toAccountId)
                .orElseThrow(() -> new RuntimeException("Receiver account not found"));

        if (fromAccount.getBalance().compareTo(amount) < 0) {
            throw new RuntimeException("Insufficient funds");
        }

        // 1. Update state
        fromAccount.setBalance(fromAccount.getBalance().subtract(amount));
        toAccount.setBalance(toAccount.getBalance().add(amount));

        accountRepository.save(fromAccount);
        accountRepository.save(toAccount);

        // 2. Build the Event Payload (In production, use a JSON library like Jackson. Keeping it simple here.)
        String eventPayload = String.format("{\"from\": \"%s\", \"to\": \"%s\", \"amount\": %s}", 
                                            fromAccountId, toAccountId, amount.toString());

        // 3. Save the Outbox Event. 
        // Because of @Transactional, if fails-account balance changes are rolled back
        OutboxEvent event = new OutboxEvent("AccountTransfer", "TransferCompleted", eventPayload);
        outboxEventRepository.save(event);
    }
}