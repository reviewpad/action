#!/bin/bash
# Copyright (C) 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
# Use of this source code is governed by a license that can be
# found in the LICENSE file.

echo "Starting semantic service..."
/app/semantic-server &

echo "Running action"
/app/action