/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

import duckdb from "duckdb"
import path from "path"

import fs from "fs"

const INIT_SCRIPT = `
-- Domain: Extensions
INSTALL spatial;
LOAD spatial;

-- Domain: Offenses
CREATE TABLE IF NOT EXISTS offenses (
    db_id INTEGER NOT NULL,
    doc_id VARCHAR,
    doc_date DATE,
    doc_source VARCHAR NOT NULL,
    record_id INTEGER NOT NULL,
    offense_id VARCHAR,
    vehicle VARCHAR,
    vehicle_country CHAR(2),
    vehicle_type VARCHAR,
    "time" TIMESTAMPTZ,
    time_year USMALLINT,
    location VARCHAR,
    display_location VARCHAR,
    description VARCHAR,
    ur INTEGER,
    error VARCHAR,
    point POINT_2D,
    h3_res1 UBIGINT,
    h3_res2 UBIGINT,
    h3_res3 UBIGINT,
    h3_res4 UBIGINT,
    h3_res5 UBIGINT,
    h3_res6 UBIGINT,
    h3_res7 UBIGINT,
    h3_res8 UBIGINT
);

ALTER TABLE offenses ADD COLUMN IF NOT EXISTS article_ids VARCHAR[];
ALTER TABLE offenses ADD COLUMN IF NOT EXISTS article_codes TINYINT[];

-- Domain: Articles & Descriptions
CREATE TABLE IF NOT EXISTS articles (
    id VARCHAR PRIMARY KEY,
    text VARCHAR NOT NULL,
    code TINYINT NOT NULL,
    title VARCHAR NOT NULL
);

CREATE SEQUENCE IF NOT EXISTS descriptions_seq;
CREATE TABLE IF NOT EXISTS descriptions (
    id INTEGER PRIMARY KEY DEFAULT nextval('descriptions_seq'),
    description VARCHAR UNIQUE NOT NULL,
    article_ids VARCHAR[],
    article_codes TINYINT[],
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Domain: Locations
CREATE SEQUENCE IF NOT EXISTS locations_seq START 1;

CREATE TABLE IF NOT EXISTS locations (
    id INTEGER PRIMARY KEY DEFAULT nextval('locations_seq'),
    db_id INTEGER NOT NULL,
    location VARCHAR NOT NULL,
    canonical_location VARCHAR,
    point POINT_2D NOT NULL,
    is_electronic BOOLEAN DEFAULT FALSE,
    geocoding_method VARCHAR NOT NULL,
    confidence VARCHAR NOT NULL,
    notes TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    h3_res1 UBIGINT,
    h3_res2 UBIGINT,
    h3_res3 UBIGINT,
    h3_res4 UBIGINT,
    h3_res5 UBIGINT,
    h3_res6 UBIGINT,
    h3_res7 UBIGINT,
    h3_res8 UBIGINT,
    UNIQUE(db_id, location)
);

-- Populate Articles
INSERT INTO articles (id, text, code, title) VALUES 
('13.3.A', 'Superar las velocidades máximas permitidas: hasta 20 km', 13, 'De las velocidades'),
('18.9.1', 'Estacionar en lugar prohibido o regulado', 18, 'Del estacionamiento'),
('21.8', 'No usar chaleco, campera 0 bandas retroreflectivas reglamentaria', 21, 'De los vehículos de dos ruedas');

-- Populate Descriptions
INSERT INTO descriptions (description, article_ids, article_codes) VALUES 
('Exceso de velocidad', ['13.3.A'], [13]),
('Estacionar en lugar prohibido', ['18.9.1'], [18]),
('No usar casco reglamentario', ['21.8'], [21]),
('Circular con deuda de patente', [], []),
('NO USAR CHALECO CAMPERA O BANDA RETRO REFLECTIVA REGLAMENTARIA', ['21.8'], [21]);

-- Populate Offenses
INSERT INTO offenses (db_id, doc_source, doc_id, doc_date, record_id, offense_id, vehicle, vehicle_country, vehicle_type, time, time_year, location, display_location, description, ur, error, point, h3_res6, h3_res7, h3_res8) VALUES
(6, 'https://www.impo.com.uy/bases/notificaciones-transito-montevideo/1234-2024', '1234/024', '2024-03-15', 1234, '1', 'AAO3197', 'UY', 'Auto', '2024-03-10 14:30:00', 2024, 'Av 18 de Julio y Rio Branco', 'Av 18 de Julio y Rio Branco', 'Exceso de velocidad', 750, NULL, ST_Point(-56.1915, -34.9055), 606990499695427583, 611494017646690303, 615997535597953023),
(6, 'https://www.impo.com.uy/bases/notificaciones-transito-montevideo/1235-2024', '1234/024', '2024-03-16', 1235, '2', 'BBX4521', 'UY', 'Auto', '2024-03-11 16:45:00', 2024, 'BV JOSE BATLLE Y ORDOÑEZ y AV DAMASO ANTONIO LARRAÑAGA', 'BV JOSE BATLLE Y ORDOÑEZ y AV DAMASO ANTONIO LARRAÑAGA', 'Estacionar en lugar prohibido', 300, NULL, ST_Point(-56.1342, -34.8857), 606990499695427583, 611494017646690303, 615997535597953023),
(6, 'https://www.impo.com.uy/bases/notificaciones-transito-montevideo/1236-2024', '1235/024', '2024-03-17', 1235, '3', 'PAV1450', 'UY', 'Moto', '2024-03-12 09:20:00', 2024, 'AV 18 DE JULIO y EJIDO', 'AV 18 DE JULIO y EJIDO', 'No usar casco reglamentario', 500, NULL, ST_Point(-56.1882, -34.9033), 606990499695427583, 611494017646690303, 615997535597953023),
(6, 'https://www.impo.com.uy/bases/notificaciones-transito-montevideo/1237-2024', '1237/024', '2024-03-18', 1237, '4', 'AAO3197', 'UY', 'Auto', '2024-03-13 11:00:00', 2024, 'AV ITALIA y PROPIOS', 'AV ITALIA y PROPIOS', 'Exceso de velocidad', 750, NULL, ST_Point(-56.1258, -34.8889), 606990499695427583, 611494017646690303, 615997535597953023),
(6, 'https://www.impo.com.uy/bases/notificaciones-transito-montevideo/1238-2024', '1238/024', '2024-03-19', 1238, '5', 'BDT956', 'UY', 'Camión', '2024-03-14 07:30:00', 2024, 'RUTA 1 KM 25', 'RUTA 1 KM 25', 'Circular con deuda de patente', 100, NULL, ST_Point(-56.2847, -34.8124), 606990499695427583, 611494017646690303, 615997535597953023),
(45, 'https://www.impo.com.uy/bases/notificaciones-transito-maldonado/567-2024', '567/024', '2024-02-20', 567, '6', 'BFM643', 'UY', 'Auto', '2024-02-15 10:15:00', 2024, 'Ruta 10, Punta del Este', 'Ruta 10, Punta del Este', 'Exceso de velocidad', 600, NULL, ST_Point(-54.9478, -34.9678), 606990499695427583, 611494017646690303, 615997535597953023),
(45, 'https://www.impo.com.uy/bases/notificaciones-transito-maldonado/568-2024', '568/024', '2024-02-21', 568, '7', 'CCY7890', 'UY', 'Auto', '2024-02-16 15:30:00', 2024, 'Av Gorlero, Punta del Este', 'Av Gorlero, Punta del Este', 'Estacionar en lugar prohibido', 300, NULL, ST_Point(-54.9483, -34.9554), 606990499695427583, 611494017646690303, 615997535597953023),
(45, 'https://www.impo.com.uy/bases/notificaciones-transito-maldonado/569-2024', '569/024', '2024-02-22', 569, '8', 'PBZ2341', 'UY', 'Moto', '2024-02-17 12:00:00', 2024, 'Ruta 9, San Carlos', 'Ruta 9, San Carlos', 'No usar casco reglamentario', 500, NULL, ST_Point(-54.9177, -34.7925), 606990499695427583, 611494017646690303, 615997535597953023),
(26, 'https://www.impo.com.uy/bases/resoluciones-transito-lavalleja/231-2024', '231/024', '2024-05-27', 90, '9', 'DDR1234', 'UY', 'Auto', '2024-05-20 10:30:00', 2024, 'RUTA 8, MINAS', 'RUTA 8, MINAS', 'Exceso de velocidad', 750, NULL, ST_Point(-55.2381, -34.3757), 606990499695427583, 611494017646690303, 615997535597953023),
(26, 'https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/14-2024', '14/024', '2024-04-16', 92, '10', 'PAV1450', 'UY', 'Moto', '2030-03-30 12:51:00', 2030, 'BALTASAR BRUN, MINAS', 'BALTASAR BRUN, MINAS', 'NO USAR CHALECO CAMPERA O BANDA RETRO REFLECTIVA REGLAMENTARIA', 500, 'la fecha 2030-03-30 09:51:00 -0300 -03 es más nueva que la fecha de publicación 2024-04-16 00:00:00 -0300 -03', ST_Point(-55.2386, -34.3833), 606990499695427583, 611494017646690303, 615997535597953023),
(40, 'https://www.impo.com.uy/bases/notificaciones-transito-canelones/100-2024', '100/024', '2024-01-15', 100, '11', 'EEF5678', 'UY', 'Auto', '2024-01-10 08:00:00', 2024, 'Ruta 5, Las Piedras', 'Ruta 5, Las Piedras', 'Circular con deuda de patente', 100, NULL, ST_Point(-56.2194, -34.7274), 606990499695427583, 611494017646690303, 615997535597953023),
(40, 'https://www.impo.com.uy/bases/notificaciones-transito-canelones/101-2024', '101/024', '2024-01-16', 101, '12', 'PCX9876', 'UY', 'Moto', '2024-01-11 14:20:00', 2024, 'Av Italia, Pando', 'Av Italia, Pando', 'No usar casco reglamentario', 500, NULL, ST_Point(-55.9583, -34.7167), 606990499695427583, 611494017646690303, 615997535597953023),
(40, 'https://www.impo.com.uy/bases/notificaciones-transito-canelones/102-2024', '102/024', '2024-01-17', 102, '13', 'AAO3197', 'UY', 'Auto', '2024-01-12 16:45:00', 2024, 'Ruta 6, Sauce', 'Ruta 6, Sauce', 'Estacionar en lugar prohibido', 300, NULL, ST_Point(-56.0631, -34.6519), 606990499695427583, 611494017646690303, 615997535597953023),
(48, 'https://www.impo.com.uy/bases/notificaciones-transito-colonia/200-2024', '200/024', '2024-06-10', 200, '14', 'FFG1111', 'UY', 'Auto', '2024-06-05 10:00:00', 2024, 'Ruta 1, Colonia del Sacramento', 'Ruta 1, Colonia del Sacramento', 'Exceso de velocidad', 750, NULL, ST_Point(-57.8397, -34.4631), 606990499695427583, 611494017646690303, 615997535597953023),
(48, 'https://www.impo.com.uy/bases/notificaciones-transito-colonia/201-2024', '201/024', '2024-06-11', 201, '15', 'BDT956', 'UY', 'Camión', '2024-06-06 07:15:00', 2024, 'Ruta 21, Carmelo', 'Ruta 21, Carmelo', 'Circular con deuda de patente', 100, NULL, NULL, NULL, NULL, NULL);
`

let dbInstance: duckdb.Database | null = null
let resolveInit: () => void
let rejectInit: (err: any) => void
const readyPromise = new Promise<void>((resolve, reject) => {
  resolveInit = resolve
  rejectInit = reject
})

export async function waitForDB(): Promise<void> {
  getDuckDB() // Ensure initialization starts
  await readyPromise
}

export function getDuckDB(): duckdb.Database {
  if (!dbInstance) {
    const dbPath = path.join(process.cwd(), "chapauy.duckdb")

    // 1. Check if DB file exists
    if (fs.existsSync(dbPath)) {
      console.log(
        `[DuckDB] Found database file at ${dbPath}. Opening in READ_ONLY mode.`
      )

      dbInstance = new duckdb.Database(dbPath, duckdb.OPEN_READONLY, (err) => {
        if (err) {
          console.error("[DuckDB] Failed to open database:", err)
          rejectInit(err)
          return
        }
        console.log("[DuckDB] Database opened successfully (READ_ONLY).")

        // Even in read-only, we need the spatial extension loaded for queries
        dbInstance?.exec("INSTALL spatial; LOAD spatial;", (err) => {
          if (err) {
            console.error("[DuckDB] Failed to load spatial extension:", err)
            // We don't reject here, maybe some queries work without it?
            // But usually it's critical. Let's log and proceed.
          }
          resolveInit()
        })
      })
    } else {
      // 2. File does not exist -> Use :memory: and initialize schema
      console.log(
        `[DuckDB] Database file not found at ${dbPath}. Using :memory: database.`
      )

      dbInstance = new duckdb.Database(":memory:", (err) => {
        if (err) {
          console.error("[DuckDB] Failed to create in-memory database:", err)
          rejectInit(err)
          return
        }

        console.log(
          "[DuckDB] In-memory database created. Initializing schema..."
        )
        if (dbInstance) {
          console.log("[InitDB] Starting initialization...")
          dbInstance.exec(INIT_SCRIPT, (err) => {
            if (err) {
              console.error("[InitDB] Error initializing database:", err)
              rejectInit(err)
              return
            }
            console.log("[InitDB] Schema initialized and mock data populated.")
            resolveInit()
          })
        }
      })
    }
  }
  return dbInstance!
}
