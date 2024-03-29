package data

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type TaxonCache struct {
	DeprecatedIds map[int64]int64        `json:"deprecated_ids"`
	Mappings      map[int64]NcbiTaxEntry `json:"mappings"`
}

func (t *TaxonCache) EntryForTaxId(taxId int64) (*NcbiTaxEntry, error) {
	entry, found := t.Mappings[taxId]
	if !found {
		newTaxId, found := t.DeprecatedIds[taxId]
		if !found {
			return nil, ErrRecordNotFound
		}
		entry, found = t.Mappings[newTaxId]
		if !found {
			return nil, ErrRecordNotFound
		}
	}
	return &entry, nil
}

type NcbiTaxEntry struct {
	TaxId        int64  `json:"tax_id"`
	Name         string `json:"name"`
	Species      string `json:"species"`
	Genus        string `json:"genus"`
	Family       string `json:"family"`
	Order        string `json:"order"`
	Class        string `json:"class"`
	Phylum       string `json:"phylum"`
	Kingdom      string `json:"kingdom"`
	Superkingdom string `json:"superkingdom"`
}

func EntryForTaxId(taxId int64) (*NcbiTaxEntry, error) {
	dumpFileName := viper.GetString("taxa.lineage")
	if dumpFileName == "" {
		return nil, ErrRecordNotFound
	}

	dumpFile, err := os.Open(dumpFileName)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(dumpFile)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.SplitN(line, "|", 11)

		for i, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed == "" {
				trimmed = "Unknown"
			}
			parts[i] = trimmed
		}

		currTaxId, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, err
		}

		if currTaxId != taxId {
			continue
		}

		species_parts := strings.SplitN(parts[2], " ", 2)

		return &NcbiTaxEntry{
			TaxId:        taxId,
			Name:         parts[1],
			Species:      species_parts[len(species_parts)-1],
			Genus:        parts[3],
			Family:       parts[4],
			Order:        parts[5],
			Class:        parts[6],
			Phylum:       parts[7],
			Kingdom:      parts[8],
			Superkingdom: parts[9],
		}, nil
	}

	err = scanner.Err()
	if err != nil {
		return nil, err
	}

	// Maybe the NCBI taxid isn't up-to-date anymore, let's try to find the new ID
	newId, err := FindMergedId(taxId)
	if err != nil {
		return nil, err
	}

	return EntryForTaxId(newId)
}

func FindMergedId(taxId int64) (int64, error) {
	mergedFileName := viper.GetString("taxa.merged")
	if mergedFileName == "" {
		return -1, ErrRecordNotFound
	}

	mergedFile, err := os.Open(mergedFileName)
	if err != nil {
		return -1, err
	}

	scanner := bufio.NewScanner(mergedFile)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.SplitN(line, "|", 3)

		oldId, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil {
			return -1, err
		}

		newId, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			return -1, err
		}

		if oldId == taxId {
			return newId, nil
		}
	}
	err = scanner.Err()
	if err != nil {
		return -1, err
	}

	return -1, ErrRecordNotFound
}
