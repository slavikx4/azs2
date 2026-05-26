BEGIN;

CREATE TYPE fuel_type AS ENUM ('АИ-92', 'АИ-95', 'АИ-98', 'ДТ');
CREATE TYPE card_status AS ENUM ('active', 'blocked', 'expired');
CREATE TYPE refuel_status AS ENUM ('pending', 'confirmed', 'rejected', 'sync_error');

CREATE TABLE vehicles (
                          id BIGSERIAL PRIMARY KEY,
                          plate_number VARCHAR(20) NOT NULL UNIQUE,
                          fuel_level NUMERIC(5,2) NOT NULL DEFAULT 0,
                          route_limit NUMERIC(10,2) NOT NULL,
                          system_card_id VARCHAR(50),
                          created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                          updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE fuel_cards (
                            id VARCHAR(50) PRIMARY KEY,
                            vehicle_id BIGINT REFERENCES vehicles(id) ON DELETE CASCADE,
                            provider_id VARCHAR(50) NOT NULL,
                            provider_card_number VARCHAR(50),
                            balance NUMERIC(10,2) NOT NULL DEFAULT 0,
                            "limit" NUMERIC(10,2) NOT NULL,
                            status card_status NOT NULL DEFAULT 'active',
                            is_active BOOLEAN NOT NULL DEFAULT true,
                            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE fuel_providers (
                                id VARCHAR(50) PRIMARY KEY,
                                name VARCHAR(100) NOT NULL,
                                is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE fuel_stations (
                               id VARCHAR(50) PRIMARY KEY,
                               provider_id VARCHAR(50) NOT NULL REFERENCES fuel_providers(id),
                               name VARCHAR(100) NOT NULL,
                               address TEXT
);

CREATE TABLE refuelings (
                            id BIGSERIAL PRIMARY KEY,
                            vehicle_id BIGINT NOT NULL REFERENCES vehicles(id),
                            system_card_id VARCHAR(50) NOT NULL,
                            provider_card_id VARCHAR(50),
                            provider_id VARCHAR(50),
                            station_id VARCHAR(50) REFERENCES fuel_stations(id),
                            station_name VARCHAR(100),
                            address TEXT,
                            liters NUMERIC(10,2) NOT NULL,
                            price_per_liter NUMERIC(8,2) NOT NULL,
                            total_amount NUMERIC(10,2) NOT NULL,
                            fuel_type fuel_type NOT NULL,
                            timestamp TIMESTAMPTZ NOT NULL,
                            status refuel_status NOT NULL DEFAULT 'confirmed',
                            provider_tx_id VARCHAR(100) UNIQUE,
                            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE telemetry_records (
                                   id BIGSERIAL PRIMARY KEY,
                                   vehicle_id BIGINT NOT NULL REFERENCES vehicles(id),
                                   fuel_level NUMERIC(5,2),
                                   mileage NUMERIC(12,2),
                                   latitude DOUBLE PRECISION,
                                   longitude DOUBLE PRECISION,
                                   speed NUMERIC(6,2),
                                   engine_on BOOLEAN,
                                   timestamp TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_vehicles_plate ON vehicles(plate_number);
CREATE INDEX idx_refuelings_vehicle_time ON refuelings(vehicle_id, timestamp DESC);
CREATE INDEX idx_telemetry_vehicle_time ON telemetry_records(vehicle_id, timestamp DESC);

CREATE OR REPLACE FUNCTION fn_update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_vehicles_updated_at
    BEFORE UPDATE ON vehicles
    FOR EACH ROW EXECUTE FUNCTION fn_update_updated_at();

INSERT INTO fuel_providers (id, name) VALUES
                                          ('gazprom', 'Газпромнефть'),
                                          ('lukoil', 'Лукойл'),
                                          ('rosneft', 'Роснефть');

INSERT INTO fuel_stations (id, provider_id, name, address) VALUES
                                                               ('st1', 'lukoil', 'Лукойл Ленина', 'ул. Ленина, 12'),
                                                               ('st2', 'gazprom', 'Газпромнефть Мира', 'пр. Мира, 45'),
                                                               ('st3', 'rosneft', 'Роснефть Гагарина', 'ул. Гагарина, 3');

INSERT INTO vehicles (id, plate_number, fuel_level, route_limit, system_card_id) VALUES
                                                                                     (101, 'А123МН77', 78, 80, '123456'),
                                                                                     (102, 'В456ОР77', 34, 100, '123457'),
                                                                                     (103, 'Е789ТС77', 25, 60, '123458'),
                                                                                     (104, 'К012УФ77', 91, 120, '123459'),
                                                                                     (105, 'М345ХЦ77', 56, 90, '123460'),
                                                                                     (106, 'Н678ЧШ99', 22, 70, '123461'),
                                                                                     (107, 'О901ЩЪ50', 42, 85, '123462'),
                                                                                     (108, 'Р234ЫЭ78', 67, 95, '123463'),
                                                                                     (109, 'С567ЮЯ16', 88, 110, '123464'),
                                                                                     (110, 'Т890ЯА02', 29, 75, '123465');

INSERT INTO fuel_cards (id, vehicle_id, provider_id, provider_card_number, balance, "limit") VALUES
                                                                                                 ('FC-G-123456', 101, 'gazprom', 'GAZ1001', 80, 80),
                                                                                                 ('FC-L-123456', 101, 'lukoil', 'LUK1001', 80, 80),
                                                                                                 ('FC-R-123456', 101, 'rosneft', 'ROS1001', 80, 80),
                                                                                                 ('FC-G-123457', 102, 'gazprom', 'GAZ1002', 100, 100),
                                                                                                 ('FC-L-123457', 102, 'lukoil', 'LUK1002', 100, 100),
                                                                                                 ('FC-R-123457', 102, 'rosneft', 'ROS1002', 100, 100),
                                                                                                 ('FC-G-123458', 103, 'gazprom', 'GAZ1003', 60, 60),
                                                                                                 ('FC-L-123458', 103, 'lukoil', 'LUK1003', 60, 60),
                                                                                                 ('FC-R-123458', 103, 'rosneft', 'ROS1003', 60, 60),
                                                                                                 ('FC-G-123459', 104, 'gazprom', 'GAZ1004', 120, 120),
                                                                                                 ('FC-L-123459', 104, 'lukoil', 'LUK1004', 120, 120),
                                                                                                 ('FC-R-123459', 104, 'rosneft', 'ROS1004', 120, 120),
                                                                                                 ('FC-G-123460', 105, 'gazprom', 'GAZ1005', 90, 90),
                                                                                                 ('FC-L-123460', 105, 'lukoil', 'LUK1005', 90, 90),
                                                                                                 ('FC-R-123460', 105, 'rosneft', 'ROS1005', 90, 90),
                                                                                                 ('FC-G-123461', 106, 'gazprom', 'GAZ1006', 70, 70),
                                                                                                 ('FC-L-123461', 106, 'lukoil', 'LUK1006', 70, 70),
                                                                                                 ('FC-R-123461', 106, 'rosneft', 'ROS1006', 70, 70),
                                                                                                 ('FC-G-123462', 107, 'gazprom', 'GAZ1007', 85, 85),
                                                                                                 ('FC-L-123462', 107, 'lukoil', 'LUK1007', 85, 85),
                                                                                                 ('FC-R-123462', 107, 'rosneft', 'ROS1007', 85, 85),
                                                                                                 ('FC-G-123463', 108, 'gazprom', 'GAZ1008', 95, 95),
                                                                                                 ('FC-L-123463', 108, 'lukoil', 'LUK1008', 95, 95),
                                                                                                 ('FC-R-123463', 108, 'rosneft', 'ROS1008', 95, 95),
                                                                                                 ('FC-G-123464', 109, 'gazprom', 'GAZ1009', 110, 110),
                                                                                                 ('FC-L-123464', 109, 'lukoil', 'LUK1009', 110, 110),
                                                                                                 ('FC-R-123464', 109, 'rosneft', 'ROS1009', 110, 110),
                                                                                                 ('FC-G-123465', 110, 'gazprom', 'GAZ1010', 75, 75),
                                                                                                 ('FC-L-123465', 110, 'lukoil', 'LUK1010', 75, 75),
                                                                                                 ('FC-R-123465', 110, 'rosneft', 'ROS1010', 75, 75);

INSERT INTO refuelings (vehicle_id, system_card_id, provider_card_id, provider_id, station_id, station_name, address, liters, price_per_liter, total_amount, fuel_type, timestamp, status, provider_tx_id) VALUES
                                                                                                                                                                                                               (101, '123456', 'LUK1001', 'lukoil', 'st1', 'Лукойл', 'ул. Ленина, 12', 42, 58.40, 2452.80, 'АИ-95', '2026-05-04 10:30:00', 'confirmed', 'tx_001'),
                                                                                                                                                                                                               (101, '123456', 'GAZ1001', 'gazprom', 'st2', 'Газпромнефть', 'пр. Мира, 45', 38, 57.90, 2200.20, 'АИ-95', '2026-05-01 14:20:00', 'confirmed', 'tx_002'),
                                                                                                                                                                                                               (102, '123457', 'LUK1002', 'lukoil', 'st1', 'Лукойл', 'пр. Мира, 102', 60, 57.80, 3468.00, 'ДТ', '2026-05-03 09:15:00', 'confirmed', 'tx_003'),
                                                                                                                                                                                                               (103, '123458', 'ROS1003', 'rosneft', 'st3', 'Роснефть', 'ул. Советская, 7', 25, 59.10, 1477.50, 'АИ-92', '2026-05-04 16:45:00', 'confirmed', 'tx_004'),
                                                                                                                                                                                                               (104, '123459', 'LUK1004', 'lukoil', 'st1', 'Лукойл', 'ш. Волоколамское, 5', 70, 57.40, 4018.00, 'АИ-98', '2026-05-03 11:00:00', 'confirmed', 'tx_005'),
                                                                                                                                                                                                               (105, '123460', 'GAZ1005', 'gazprom', 'st2', 'Газпромнефть', 'ул. Тверская, 25', 48, 58.60, 2812.80, 'ДТ', '2026-05-02 13:30:00', 'confirmed', 'tx_006'),
                                                                                                                                                                                                               (106, '123461', 'ROS1006', 'rosneft', 'st3', 'Роснефть', 'ул. Садовая, 3', 20, 59.50, 1190.00, 'АИ-92', '2026-05-04 08:00:00', 'confirmed', 'tx_007'),
                                                                                                                                                                                                               (107, '123462', 'LUK1007', 'lukoil', 'st1', 'Лукойл', 'пр. Космонавтов, 5', 45, 58.00, 2610.00, 'АИ-95', '2026-05-03 15:20:00', 'confirmed', 'tx_008'),
                                                                                                                                                                                                               (108, '123463', 'GAZ1008', 'gazprom', 'st2', 'Газпромнефть', 'ул. Тверская, 25', 60, 57.60, 3456.00, 'ДТ', '2026-05-04 12:00:00', 'confirmed', 'tx_009'),
                                                                                                                                                                                                               (109, '123464', 'LUK1009', 'lukoil', 'st1', 'Лукойл', 'пр. Космонавтов, 5', 70, 57.50, 4025.00, 'АИ-98', '2026-05-02 10:00:00', 'confirmed', 'tx_010'),
                                                                                                                                                                                                               (110, '123465', 'GAZ1010', 'gazprom', 'st2', 'Газпромнефть', 'ул. Профсоюзная, 44', 25, 59.20, 1480.00, 'АИ-92', '2026-05-04 17:30:00', 'confirmed', 'tx_011');

INSERT INTO telemetry_records (vehicle_id, fuel_level, mileage, latitude, longitude, speed, engine_on, timestamp) VALUES
                                                                                                                      (101, 78, 12500, 55.7512, 37.6184, 65, true, NOW()),
                                                                                                                      (102, 34, 34200, 55.7523, 37.6195, 72, true, NOW()),
                                                                                                                      (103, 25, 8900, 55.7534, 37.6206, 0, false, NOW()),
                                                                                                                      (104, 91, 45600, 55.7545, 37.6217, 88, true, NOW()),
                                                                                                                      (105, 56, 23400, 55.7556, 37.6228, 45, true, NOW()),
                                                                                                                      (106, 22, 12300, 55.7567, 37.6239, 0, false, NOW()),
                                                                                                                      (107, 42, 56700, 55.7578, 37.6250, 52, true, NOW()),
                                                                                                                      (108, 67, 78900, 55.7589, 37.6261, 60, true, NOW()),
                                                                                                                      (109, 88, 34500, 55.7590, 37.6272, 70, true, NOW()),
                                                                                                                      (110, 29, 67800, 55.7601, 37.6283, 0, false, NOW());

COMMIT;