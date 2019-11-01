php-chtseg
===

The PHP API for chtseg is done by using 

```
github.com/kitech/php-go
```
please see its readme for installation.

Two things that are not noted on its document:

1. Besides `go install` to install `php-go`, it seems to be necessary to run `make all` in the `php-go` directory after running `go get github.com/kitech/php-go`
2. The `php-go` is tested on 64bit platform like `aarch64`. It should be OK too in `x86_64` platform. However, there are errors during compiling at `arm` platform. Maybe someone can solve it, or if I have time to.

This API is tested on both **apache2 with mod_php(7.2)** and **ngnix with fpm_php(7.2)**. However, there is a problem on mysql DB. Currently, only sqlite3 DB can be used on PHP API.

After installing the `php-go` mentioned above, simply run:

```
$ sh install.sh
```
and if will produce the php extension and install it to your php installation automatically if you provide the sudo password. Restart your apache2 or ngnix/fpm_php, all will be set.

## Usage

Please see the `test.php` on how to use it. The main and only API is

```
$d = new PHPseg()
```
and 

```
$d->Getchtseg( "sqlite3", "db file path", "sentence to be processed")
```

the output will be a json string like

```
array (
  'json' => '{"OrigInput":"現貨附發票 Raspberry Pi 樹莓派專用 USB電腦遙控器","UnsymInput":"現貨附發票 Raspberry Pi 樹莓派專用 USB電腦遙控器","Score":3.1817536180569292,"SegItems":["現貨","附發票","Raspberry","Pi","樹莓派","專用","USB","電腦","遙控器"],"NumWords":9,"Guessed":{"專用":5.179810222878795,"樹莓派":7.863405189790678,"現貨":6.541704023284288,"遙控器":9.199591795319709,"附發票":4.903089986991943,"電腦":6.843372950967203}}',
  'error' => '',
)
```
you can use php code to parse the returned json value or see the error string if there is an error.

Also you can run 

```
make chtseg-test
```
to see the debug output of `test.php`

