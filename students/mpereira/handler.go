package urlshort

import (
	json "encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/boltdb/bolt"
	yamlV2 "gopkg.in/yaml.v2"
)

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, db *bolt.DB, fallback http.Handler) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if db != nil {

			err := db.View(func(tx *bolt.Tx) error {

				b := tx.Bucket([]byte("PathRedirect"))

				v := b.Get([]byte(r.URL.Path))

				fmt.Printf("just read: %s", string(v))

				http.Redirect(w, r, string(v), http.StatusPermanentRedirect)

				return nil
			})

			if err != nil {
				fallback.ServeHTTP(w, r)
			}

		} else {
			if val, ok := pathsToUrls[r.URL.Path]; ok {
				http.Redirect(w, r, val, http.StatusPermanentRedirect)
			} else {
				fallback.ServeHTTP(w, r)
			}
		}
	}
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//     - path: /some-path
//       url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func YAMLHandler(yaml []byte, yamlFilePath string, fallback http.Handler) (http.HandlerFunc, error) {

	var parsedYaml []pathUrlObj
	var err error

	if yamlFilePath != "" {
		yamlFile, redErr := ioutil.ReadFile(yamlFilePath)

		if redErr != nil {
			return nil, redErr
		}

		parsedYaml, err = parseYAML(yamlFile)

	} else {

		parsedYaml, err = parseYAML(yaml)

	}

	if err != nil {
		return nil, err
	}

	pathMap := buildMap(parsedYaml)

	return MapHandler(pathMap, nil, fallback), nil
}

// JSONHandler will parse the provided JSON and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// JSON is expected to be in the format:
//
// {
//     "path" : "/some-path",
//     "url": "https://www.some-url.com/demo"
// }
//
// The only errors that can be returned all related to having
// invalid JSON data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func JSONHandler(json []byte, jsonFilePath string, db *bolt.DB, fallback http.Handler) (http.HandlerFunc, error) {

	var parsedJson pathUrlObjJson
	var err error

	if jsonFilePath != "" {
		jsonFile, redErr := ioutil.ReadFile(jsonFilePath)

		if redErr != nil {
			return nil, redErr
		}

		parsedJson, err = parseJSON(jsonFile)

	} else {

		parsedJson, err = parseJSON(json)

	}

	if err != nil {
		return nil, err
	}

	pathMap := buildMapFromJson(parsedJson)

	err = db.Update(func(tx *bolt.Tx) error {

		b, err := tx.CreateBucketIfNotExists([]byte("PathRedirect"))

		for k, v := range pathMap {

			err = b.Put([]byte(k), []byte(v))

		}

		return err
	})

	return MapHandler(pathMap, db, fallback), nil
}

type pathUrlObj struct {
	Path string `yaml:"path"`
	Url  string `yaml:"url"`
}

type pathUrlObjJson struct {
	PathURL []pathUrlUnit `json:"PathUrl"`
}

type pathUrlUnit struct {
	Path string `json:"path"`
	URL  string `json:"url"`
}

func parseYAML(yaml []byte) ([]pathUrlObj, error) {

	var pathList []pathUrlObj

	err := yamlV2.Unmarshal(yaml, &pathList)

	if err != nil {
		return pathList, err
	}
	return pathList, nil
}

func parseJSON(jsonFile []byte) (pathUrlObjJson, error) {

	var pathList pathUrlObjJson

	err := json.Unmarshal(jsonFile, &pathList)

	if err != nil {
		return pathList, err
	}
	return pathList, nil
}

func buildMap(pathUrlObj []pathUrlObj) map[string]string {

	pathMap := make(map[string]string, len(pathUrlObj))

	for _, path := range pathUrlObj {
		pathMap[path.Path] = path.Url
	}

	return pathMap
}

func buildMapFromJson(pathUrlObj pathUrlObjJson) map[string]string {

	pathMap := make(map[string]string, len(pathUrlObj.PathURL))

	for _, path := range pathUrlObj.PathURL {
		pathMap[path.Path] = path.URL
	}

	return pathMap
}
