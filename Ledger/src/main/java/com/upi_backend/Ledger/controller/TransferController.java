package com.upi_backend.Ledger.controller;

import com.upi_backend.Ledger.service.TransferService;
import org.springframework.web.bind.annotation.*;
import java.math.BigDecimal;

@RestController
@RequestMapping("/api")
public class TransferController {

    private final TransferService transferService;

    public TransferController(TransferService transferService) {
        this.transferService = transferService;
    }

    @PostMapping("/transfer")
    public String transfer(@RequestParam String from, @RequestParam String to, @RequestParam BigDecimal amount) {
        try {
            transferService.transferMoney(from, to, amount);
            return "Transfer successful!";
        } catch (Exception e) {
            return "Transfer failed: " + e.getMessage();
        }
    }
}