package com.upi_backend.Ledger.model;

import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import java.math.BigDecimal;

@Entity
public class Account {
    @Id
    private String accountId;
    private BigDecimal balance;

    public Account() {} 

    public Account(String accountId, BigDecimal balance) {
        this.accountId = accountId;
        this.balance = balance;
    }

    public String getAccountId() { return accountId; }
    public BigDecimal getBalance() { return balance; }
    
    public void setBalance(BigDecimal balance) { this.balance = balance; }
}