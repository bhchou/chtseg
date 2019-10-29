<?php
function test_classes() {
    $d = new PHPseg();
    $r4 = $d->Getchtseg("sqlite3", "/tmp/chtseg.db", "現貨附發票 Raspberry Pi 樹莓派專用 USB電腦遙控器");
    print_r("assoc JSON result: " .var_export($r4, true));
    // test for executing scope destruct and execute finished scope destruct
    if (rand() % 2 == 1) {
        $d = null;
    }
}
test_classes();
?>
