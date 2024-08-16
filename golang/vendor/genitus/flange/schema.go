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

package flange

// endpoint schema.
const ENDPOINT_SCHEMA = `
{
  "type" : "record",
  "name" : "Endpoint",
  "namespace" : "org.genitus.gasket",
  "fields" : [ {
    "name" : "serviceName",
    "type" : "string"
  }, {
    "name" : "ip",
    "type" : "int"
  }, {
    "name" : "port",
    "type" : "int"
  } ]
}`

// annotation schema.
const ANNOTATION_SCHEMA = `
{
  "type" : "record",
  "name" : "Annotation",
  "namespace" : "org.genitus.gasket",
  "fields" : [ {
    "name" : "timestamp",
    "type" : "long"
  }, {
    "name" : "value",
    "type" : "string"
  }, {
    "name" : "endpoint",
    "type" : {
      "type" : "record",
      "name" : "Endpoint",
      "fields" : [ {
        "name" : "serviceName",
        "type" : "string"
      }, {
        "name" : "ip",
        "type" : "int"
      }, {
        "name" : "port",
        "type" : "int"
      } ]
    }
  } ]
}`

// tag schema
const TAG_SCHEMA = `
{
  "type" : "record",
  "name" : "Tag",
  "namespace" : "org.genitus.gasket",
  "fields" : [ {
    "name" : "key",
    "type" : "string"
  }, {
    "name" : "value",
    "type" : "string"
  }, {
    "name" : "endpoint",
    "type" : {
      "type" : "record",
      "name" : "Endpoint",
      "fields" : [ {
        "name" : "serviceName",
        "type" : "string"
      }, {
        "name" : "ip",
        "type" : "int"
      }, {
        "name" : "port",
        "type" : "int"
      } ]
    }
  } ]
}`

// span schema.
const SPAN_SCHEMA = `
{
  "type" : "record",
  "name" : "Span",
  "namespace" : "org.genitus.gasket",
  "fields" : [ {
    "name" : "traceId",
    "type" : "string"
  }, {
    "name" : "name",
    "type" : "string"
  }, {
    "name" : "id",
    "type" : "string"
  }, {
    "name" : "timestamp",
    "type" : "long"
  }, {
    "name" : "duration",
    "type" : "long"
  }, {
    "name" : "annotations",
    "type" : {
      "type" : "array",
      "items" : {
        "type" : "record",
        "name" : "Annotation",
        "fields" : [ {
          "name" : "timestamp",
          "type" : "long"
        }, {
          "name" : "value",
          "type" : "string"
        }, {
          "name" : "endpoint",
          "type" : {
            "type" : "record",
            "name" : "Endpoint",
            "fields" : [ {
              "name" : "serviceName",
              "type" : "string"
            }, {
              "name" : "ip",
              "type" : "int"
            }, {
              "name" : "port",
              "type" : "int"
            } ]
          }
        } ]
      },
      "java-class" : "java.util.List"
    }
  }, {
    "name" : "tags",
    "type" : {
      "type" : "array",
      "items" : {
        "type" : "record",
        "name" : "Tag",
        "fields" : [ {
          "name" : "key",
          "type" : "string"
        }, {
          "name" : "value",
          "type" : "string"
        }, {
          "name" : "endpoint",
          "type" : "Endpoint"
        } ]
      },
      "java-class" : "java.util.List"
    }
  } ]
}`
