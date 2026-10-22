# Smoke tests for the Tasks/Todo API
# Works in Windows PowerShell 5.1 and PowerShell 7+.
#
# 1. Start the server in one terminal (JWT_SECRET must be set).
# 2. Run this script in another terminal:   .\test-api.ps1
#
# Change $Resource to "todos" if your routes use /todos.
# Every run registers fresh users, so you do not need to clean the database.

$base     = "http://localhost:8080"
$Resource = "tasks"
$stamp    = Get-Date -Format "HHmmss"

$script:pass = 0
$script:fail = 0

# ---------- helpers ----------

function Call {
    param(
        [string]$Method,
        [string]$Path,
        $Body = $null,
        [string]$Token = $null
    )
    $params = @{
        Method          = $Method
        Uri             = "$base$Path"
        ContentType     = "application/json"
        UseBasicParsing = $true
    }
    if ($Token) { $params.Headers = @{ Authorization = "Bearer $Token" } }
    if ($null -ne $Body) {
        if ($Body -is [string]) { $params.Body = $Body }
        else { $params.Body = $Body | ConvertTo-Json -Compress }
    }

    try {
        $r    = Invoke-WebRequest @params
        $code = [int]$r.StatusCode
        $text = $r.Content
    } catch {
        if ($_.Exception.Response) { $code = [int]$_.Exception.Response.StatusCode } else { $code = 0 }
        $text = $_.ErrorDetails.Message
        if (-not $text) { $text = $_.Exception.Message }
    }
    [pscustomobject]@{ Method = $Method; Path = $Path; Code = $code; Body = $text }
}

function ToObj($res) {
    try { $res.Body | ConvertFrom-Json } catch { $null }
}

function Cnt($x) {
    if ($null -eq $x) { 0 } else { @($x).Count }
}

function Assert {
    param([string]$Title, [bool]$Condition, [string]$Detail = "")
    if ($Condition) {
        $script:pass++
        Write-Host ("[ OK ] {0}" -f $Title) -ForegroundColor Green
    } else {
        $script:fail++
        Write-Host ("[FAIL] {0}" -f $Title) -ForegroundColor Red
        if ($Detail) { Write-Host ("       {0}" -f $Detail) -ForegroundColor DarkGray }
    }
}

function Check {
    param([string]$Title, $Res, [int]$Expected)
    Assert ("{0}  ({1} {2} -> {3})" -f $Title, $Res.Method, $Res.Path, $Res.Code) `
           ($Res.Code -eq $Expected) `
           ("expected {0}, got {1}. Body: {2}" -f $Expected, $Res.Code, $Res.Body)
}

function Section($name) {
    Write-Host ""
    Write-Host "=== $name ===" -ForegroundColor Cyan
}

# ---------- 1. registration ----------

Section "1. Registration"

$userA = @{ name = "Alice $stamp"; email = "alice$stamp@test.com"; password = "password123" }
$userB = @{ name = "Bob $stamp";   email = "bob$stamp@test.com";   password = "password123" }

$r = Call POST "/register" $userA
Check "Register user A" $r 201
$tokenA = (ToObj $r).token
Assert "Register response contains a token (A)" ([bool]$tokenA) $r.Body

$r = Call POST "/register" $userB
Check "Register user B" $r 201
$tokenB = (ToObj $r).token
Assert "Register response contains a token (B)" ([bool]$tokenB) $r.Body

$dup = @{ name = "Other $stamp"; email = $userA.email; password = "password123" }
Check "Duplicate email is rejected" (Call POST "/register" $dup) 409

$short = @{ name = "Short $stamp"; email = "short$stamp@test.com"; password = "123" }
Check "Too short password is rejected" (Call POST "/register" $short) 400

$noName = @{ name = ""; email = "noname$stamp@test.com"; password = "password123" }
Check "Empty name is rejected" (Call POST "/register" $noName) 400

Check "Malformed JSON is rejected" (Call POST "/register" '{"name": "x", ') 400

# ---------- 2. login ----------

Section "2. Login"

$r = Call POST "/login" @{ email = $userA.email; password = $userA.password }
Check "Login with correct credentials" $r 200
Assert "Login response contains a token" ([bool](ToObj $r).token) $r.Body
$tokenA = (ToObj $r).token

Check "Wrong password gives 401" (Call POST "/login" @{ email = $userA.email; password = "wrong-password" }) 401
Check "Unknown email gives 401" (Call POST "/login" @{ email = "nobody$stamp@test.com"; password = "password123" }) 401
Check "Empty login body is rejected" (Call POST "/login" "{}") 400

# ---------- 3. authentication ----------

Section "3. Authentication (token required)"

Check "List without token" (Call GET "/$Resource") 401
Check "List with garbage token" (Call GET "/$Resource" $null "abc.def.ghi") 401
Check "Create without token" (Call POST "/$Resource" @{ title = "x"; description = "y" }) 401
Check "Update without token" (Call PUT "/$Resource/1" @{ title = "x"; description = "y" }) 401
Check "Delete without token" (Call DELETE "/$Resource/1") 401

# ---------- 4. create ----------

Section "4. Create"

$r = Call POST "/$Resource" @{ title = "Buy groceries"; description = "Buy milk, eggs, and bread" } $tokenA
Check "A creates a task" $r 201
$task = ToObj $r
$taskId = $task.id
Assert "Created task has an id and the right title" (($taskId -gt 0) -and ($task.title -eq "Buy groceries")) $r.Body

Check "Empty title is rejected" (Call POST "/$Resource" @{ title = ""; description = "y" } $tokenA) 400
Check "Malformed JSON is rejected" (Call POST "/$Resource" '{"title": ' $tokenA) 400

# ---------- 5. isolation between users ----------

Section "5. Users see only their own tasks"

$r = Call GET "/$Resource" $null $tokenA
Check "A lists tasks" $r 200
Assert "A sees at least one task" ((Cnt (ToObj $r).data) -ge 1) $r.Body

$r = Call GET "/$Resource" $null $tokenB
Check "B lists tasks" $r 200
Assert "B sees no tasks" ((Cnt (ToObj $r).data) -eq 0) $r.Body

Check "A reads own task" (Call GET "/$Resource/$taskId" $null $tokenA) 200
Check "B reads A's task (hidden)" (Call GET "/$Resource/$taskId" $null $tokenB) 404

# ---------- 6. update ----------

Section "6. Update (PUT)"

$upd = @{ title = "Buy groceries"; description = "Buy milk, eggs, bread, and cheese" }

$r = Call PUT "/$Resource/$taskId" $upd $tokenA
Check "A updates own task" $r 200
Assert "Description was updated" ((ToObj $r).description -eq $upd.description) $r.Body

Check "B updates A's task (forbidden)" (Call PUT "/$Resource/$taskId" $upd $tokenB) 403
Check "Update non-existent task" (Call PUT "/$Resource/999999" $upd $tokenA) 404
Check "Update with invalid id" (Call PUT "/$Resource/abc" $upd $tokenA) 400
Check "Update with invalid body" (Call PUT "/$Resource/$taskId" @{ title = ""; description = "" } $tokenA) 400

# ---------- 7. delete ----------

Section "7. Delete"

Check "B deletes A's task (forbidden)" (Call DELETE "/$Resource/$taskId" $null $tokenB) 403
Check "Delete non-existent task" (Call DELETE "/$Resource/999999" $null $tokenA) 404
Check "Delete with invalid id" (Call DELETE "/$Resource/abc" $null $tokenA) 400

$r = Call DELETE "/$Resource/$taskId" $null $tokenA
Check "A deletes own task" $r 204
Assert "204 response has an empty body" ([string]::IsNullOrEmpty($r.Body)) $r.Body

Check "Deleted task is gone" (Call GET "/$Resource/$taskId" $null $tokenA) 404
Check "Deleting twice gives 404" (Call DELETE "/$Resource/$taskId" $null $tokenA) 404

# ---------- 8. pagination ----------

Section "8. Pagination (A has 5 tasks now)"

1..5 | ForEach-Object {
    $null = Call POST "/$Resource" @{ title = "Task $_"; description = "Description $_" } $tokenA
}

$r = Call GET "/${Resource}?page=1&limit=2" $null $tokenA
Check "Page 1, limit 2" $r 200
$j = ToObj $r
Assert "Page 1 has 2 items" ((Cnt $j.data) -eq 2) $r.Body
Assert "Response has page=1, limit=2, total=5" (($j.page -eq 1) -and ($j.limit -eq 2) -and ($j.total -eq 5)) $r.Body

$r = Call GET "/${Resource}?page=3&limit=2" $null $tokenA
Check "Page 3, limit 2" $r 200
$j = ToObj $r
Assert "Last page has the 1 remaining item" ((Cnt $j.data) -eq 1) $r.Body
Assert "Total is still 5" ($j.total -eq 5) $r.Body

$r = Call GET "/${Resource}?page=4&limit=2" $null $tokenA
Check "Page beyond the end" $r 200
Assert "Page beyond the end is empty" ((Cnt (ToObj $r).data) -eq 0) $r.Body

$r = Call GET "/$Resource" $null $tokenA
Check "No pagination parameters (defaults)" $r 200
$j = ToObj $r
Assert "Defaults are page=1, limit=10" (($j.page -eq 1) -and ($j.limit -eq 10)) $r.Body

$r = Call GET "/$Resource" $null $tokenB
Assert "B's total is 0 (counts only own tasks)" ((ToObj $r).total -eq 0) $r.Body

Check "page=0 is rejected" (Call GET "/${Resource}?page=0" $null $tokenA) 400
Check "page=abc is rejected" (Call GET "/${Resource}?page=abc" $null $tokenA) 400
Check "limit=abc is rejected" (Call GET "/${Resource}?limit=abc" $null $tokenA) 400
Check "limit=0 is rejected" (Call GET "/${Resource}?limit=0" $null $tokenA) 400

$r = Call GET "/${Resource}?page=1&limit=1000" $null $tokenA
if ($r.Code -eq 200) {
    Write-Host ("[INFO] limit=1000 was accepted and capped to limit={0}" -f (ToObj $r).limit) -ForegroundColor Yellow
} else {
    Write-Host ("[INFO] limit=1000 was rejected with {0} (a cap or a 400 are both fine)" -f $r.Code) -ForegroundColor Yellow
}

# ---------- summary ----------

Write-Host ""
Write-Host "=== Summary ===" -ForegroundColor Cyan
Write-Host ("Passed: {0}" -f $script:pass) -ForegroundColor Green
if ($script:fail -gt 0) {
    Write-Host ("Failed: {0}" -f $script:fail) -ForegroundColor Red
    exit 1
} else {
    Write-Host "All checks passed." -ForegroundColor Green
}
