/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package quiver

// tag schema
const MEDIUM_SCHEMA = `
{
  "type" : "record",
  "name" : "Medium",
  "namespace" : "org.genitus.shaft.v2.event",
  "fields" : [ {
    "name" : "key",
    "type" : "string"
  }, {
    "name" : "type",
    "type" : "string"
  }, {
    "name" : "header",
    "type" : "string"
  }, {
    "name" : "data",
    "type" : {
      "type" : "bytes",
      "java-class" : "[B"
    }
  } ]
}`

// span schema.
const EVENT_SCHEMA = `
{
  "type" : "record",
  "name" : "Event",
  "namespace" : "org.genitus.shaft.v2.event",
  "fields" : [ {
    "name" : "type",
    "type" : "int"
  }, {
    "name" : "sid",
    "type" : "string"
  }, {
    "name" : "uid",
    "type" : "string"
  }, {
    "name" : "syncid",
    "type" : "int"
  }, {
    "name" : "sub",
    "type" : "string"
  }, {
    "name" : "timestamp",
    "type" : "long"
  }, {
    "name" : "name",
    "type" : "string"
  }, {
    "name" : "endpoint",
    "type" : "string"
  }, {
    "name" : "tags",
    "type" : {
      "type" : "map",
      "values" : "string"
    }
  }, {
    "name" : "outputs",
    "type" : {
      "type" : "map",
				"values" : {
					"type" : "array",
					"items" : "string",
					"java-class" : "java.util.List"
				}
    }
  }, {
    "name" : "descs",
    "type" : {
      "type" : "array",
      "items" : "string",
      "java-class" : "java.util.List"
    }
  }, {
    "name" : "media",
    "type" : {
      "type" : "array",
      "items" : {
        "type" : "record",
        "name" : "Medium",
        "fields" : [ {
          "name" : "key",
          "type" : "string"
        }, {
          "name" : "type",
          "type" : "string"
        }, {
          "name" : "header",
          "type" : "string"
        }, {
          "name" : "data",
          "type" : {
            "type" : "bytes",
            "java-class" : "[B"
          }
        } ]
      },
      "java-class" : "java.util.List"
    }
  } ]
}`
