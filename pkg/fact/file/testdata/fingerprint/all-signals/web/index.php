<?php

use Drupal\Core\DrupalKernel;

$autoloader = require_once 'autoload.php';
$kernel = new DrupalKernel('prod', $autoloader);