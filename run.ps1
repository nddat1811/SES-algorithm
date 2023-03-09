rm .\static\logs\*.log

$n_processes = Read-Host "Enter the number of processes"

if (![int]::TryParse($n_processes, [ref]$null)) {
  Write-Host "Invalid input. Please enter an integer."
  Exit
}

$processes_arr = @()
for ($i = 0; $i -lt $n_processes; $i++) {
  Write-Host "Starting process $i."
  Start-Process go -ArgumentList "run main.go $n_processes $i"
  $processes_arr += $pid
}
