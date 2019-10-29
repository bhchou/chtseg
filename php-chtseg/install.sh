#!/bin/bash
make chtseg
SO=php-chtseg.so
if [ -e $SO ]; then
    sudo make chtseg-install
fi
PHPETC=$(echo `php -i | grep "Configuration File (php.ini)"` | gawk 'BEGIN {FS="/"}; {print "/"$2"/"$3"/"$4}' )
INIDIR=$PHPETC/mods-available
if [ ! -d $INIDIR ]; then
    echo "The php installation is not as what I have known, please copy necessary files manually"
    exit
else
    if [ -e $INIDIR/php-chtseg.ini ]; then
        sudo rm $INIDIR/php-chtseg.ini
    fi
    sudo cp php-chtseg.ini $INIDIR/
fi
if [ -d $PHPETC/apache2/conf.d ]; then
    if [ -e $PHPETC/apache2/conf.d/30-chtseg.ini ]; then
        sudo rm $PHPETC/apache2/conf.d/30-chtseg.ini
    fi
    sudo ln -sf  $INIDIR/php-chtseg.ini $PHPETC/apache2/conf.d/30-chtseg.ini
fi
if [ -d $PHPETC/fpm/conf.d ]; then
    if [ -e $PHPETC/fpm/conf.d/30-chtseg.ini ]; then
        sudo rm $PHPETC/fpm/conf.d/30-chtseg.ini
    fi
    sudo ln -sf $INIDIR/php-chtseg.ini $PHPETC/fpm/conf.d/30-chtseg.ini
fi
echo "---Please restart your apache/ngnix/php service to enable the extension----"
echo "If there is an error above, please check the directory denoted in Makefile and your php install directories"
echo "then do the install process manually"


