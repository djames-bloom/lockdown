# Lockdown

Toolkit of utilities to detect possible VM imaging, rollbacks and snapshots

## Current
- CPU timing inconsistencies
- Memory layout analysis
- Monotonic clock analysis
- Entropy PRNG pattern analysis

## Planned
- Stack corruption
- Sync state
- CPUID latency
- Restricted registers

## Potential
These are hardware dependent and are not reliable on any environment bar bare
metal either due to not being available at all, or virtualization prioritizing
over hardware (e.g. EC2 vTPM)

- SGX enclaves (Intel CPU required and enabled in UEFI)
- SEV/SNP (AMD CPU required)
- TPM Root of Trust (Can be falsified with virtual TPMs)
- HSM (hardware SOC required)
- UEFI boot chain verification (secure boot required)
- DMA detection (IOMMU required)
